package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sns"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// Records greater than 256 KB in size cannot be
// put into an SNS topic.
const sendAWSSNSMessageSizeLimit = 1024 * 1024 * 256

// errSendAWSSNSMessageSizeLimit is returned when data exceeds the SNS msg
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendAWSSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSNSConfig struct {
	Object        iconfig.Object  `json:"object"`
	Batch         iconfig.Batch   `json:"batch"`
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	AWS   iconfig.AWS   `json:"aws"`
	Retry iconfig.Retry `json:"retry"`

	// ARN is the AWS SNS topic ARN that messages are sent to.
	ARN string `json:"arn"`
}

func (c *sendAWSSNSConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSSNSConfig) Validate() error {
	if c.ARN == "" {
		return fmt.Errorf("topic: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSSNS(_ context.Context, cfg config.Config) (*sendAWSSNS, error) {
	conf := sendAWSSNSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
	}

	tf := sendAWSSNS{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		RoleARN:    conf.AWS.RoleARN,
		MaxRetries: conf.Retry.Count,
	})

	agg, err := aggregate.New(aggregate.Config{
		// SQS limits batch operations to 10 messages.
		Count: 10,
		// SNS limits batch operations to 256 KB.
		Size:     sendAWSSNSMessageSizeLimit,
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
				return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendAWSSNS struct {
	conf sendAWSSNSConfig

	// client is safe for concurrent use.
	client sns.API

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSSNS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSSNSMessageSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", errSendAWSSNSMessageSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", errSendBatchMisconfigured)
	}
	return []*message.Message{msg}, nil
}

func (tf *sendAWSSNS) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSSNS) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	if _, err := tf.client.PublishBatch(ctx, tf.conf.ARN, data); err != nil {
		return err
	}

	return nil
}
