package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// Records greater than 256 KB in size cannot be
// put into an SNS topic.
const sendAWSSNSMessageSizeLimit = 1024 * 1024 * 256

// errSendAWSSNSMessageSizeLimit is returned when data exceeds the SNS msg
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendAWSSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSNSConfig struct {
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *sendAWSSNSConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSSNSConfig) Validate() error {
	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSSNS(ctx context.Context, cfg config.Config) (*sendAWSSNS, error) {
	conf := sendAWSSNSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_sns: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_sns"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSSNS{
		conf: conf,
	}

	awsCfg, err := iconfig.NewAWS(ctx, conf.AWS)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = sns.NewFromConfig(awsCfg)

	// SNS limits batch operations to 10 messages.
	count := 10
	if conf.Batch.Count > 0 && conf.Batch.Count <= count {
		count = conf.Batch.Count
	}

	// SNS limits batch operations to 256 KB.
	size := sendAWSSNSMessageSizeLimit
	if conf.Batch.Size > 0 && conf.Batch.Size <= size {
		size = conf.Batch.Size
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    count,
		Size:     size,
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

type sendAWSSNS struct {
	conf   sendAWSSNSConfig
	client *sns.Client

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
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSSNSMessageSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSSNSMessageSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
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

	ctx = context.WithoutCancel(ctx)
	return tf.sendMessages(ctx, data)
}

func (tf *sendAWSSNS) sendMessages(ctx context.Context, data [][]byte) error {
	mgid := uuid.New().String()

	entries := make([]types.PublishBatchRequestEntry, 0, len(data))
	for idx, d := range data {
		entry := types.PublishBatchRequestEntry{
			Id:      aws.String(strconv.Itoa(idx)),
			Message: aws.String(string(d)),
		}

		if strings.HasSuffix(tf.conf.AWS.ARN, ".fifo") {
			entry.MessageGroupId = aws.String(mgid)
		}

		entries = append(entries, entry)
	}

	ctx = context.WithoutCancel(ctx)
	resp, err := tf.client.PublishBatch(ctx, &sns.PublishBatchInput{
		PublishBatchRequestEntries: entries,
		TopicArn:                   aws.String(tf.conf.AWS.ARN),
	})
	if err != nil {
		return err
	}

	if resp.Failed != nil {
		var retry [][]byte
		for _, r := range resp.Failed {
			idx, err := strconv.Atoi(aws.StringValue(r.Id))
			if err != nil {
				return err
			}

			retry = append(retry, data[idx])
		}

		if len(retry) > 0 {
			return tf.sendMessages(ctx, retry)
		}
	}

	return nil
}
