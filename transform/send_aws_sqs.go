package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sqs"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// Records greater than 256 KB in size cannot be
// put into an SQS queue.
const sendSQSMessageSizeLimit = 1024 * 1024 * 256

// errSendSQSMessageSizeLimit is returned when data exceeds the SQS msg
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSQSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSQSConfig struct {
	// ARN is the AWS SNS topic ARN that messages are sent to.
	ARN string `json:"arn"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`
}

func (c *sendAWSSQSConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSSQSConfig) Validate() error {
	if c.ARN == "" {
		return fmt.Errorf("arn: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSSQS(_ context.Context, cfg config.Config) (*sendAWSSQS, error) {
	conf := sendAWSSQSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_sqs: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_sqs"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	// arn:aws:sqs:region:account_id:queue_name
	arn := strings.Split(conf.ARN, ":")
	tf := sendAWSSQS{
		conf: conf,
		queueURL: fmt.Sprintf(
			"https://sqs.%s.amazonaws.com/%s/%s",
			arn[3],
			arn[4],
			arn[5],
		),
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:          conf.AWS.Region,
		RoleARN:         conf.AWS.RoleARN,
		MaxRetries:      conf.Retry.Count,
		RetryableErrors: conf.Retry.ErrorMessages,
	})

	agg, err := aggregate.New(aggregate.Config{
		// SQS limits batch operations to 10 messages.
		Count: 10,
		// SQS limits batch operations to 256 KB.
		Size:     sendSQSMessageSizeLimit,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, err
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendAWSSQS struct {
	conf     sendAWSSQSConfig
	queueURL string

	// client is safe for concurrent use.
	client sqs.API

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSSQS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendSQSMessageSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendSQSMessageSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}
	return []*message.Message{msg}, nil
}

func (tf *sendAWSSQS) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSSQS) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	if _, err := tf.client.SendMessageBatch(ctx, tf.queueURL, data); err != nil {
		return err
	}

	return nil
}
