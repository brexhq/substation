package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/firehose"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// Records greater than 1000 KiB in size cannot be put into Kinesis Firehose.
const sendAWSKinesisDataFirehoseMessageSizeLimit = 1024 * 1000

// errSendAWSKinesisDataFirehoseRecordSizeLimit is returned when data exceeds the
// Kinesis Firehose record size limit. If this error occurs,
// then drop or reduce the size of the data before attempting to
// send it to Kinesis Firehose.
var errSendAWSKinesisDataFirehoseRecordSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSKinesisDataFirehoseConfig struct {
	// StreamName is the Firehose Delivery Stream that records are sent to.
	StreamName string `json:"stream_name"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`
}

func (c *sendAWSKinesisDataFirehoseConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSKinesisDataFirehoseConfig) Validate() error {
	if c.StreamName == "" {
		return fmt.Errorf("stream_name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSKinesisDataFirehose(_ context.Context, cfg config.Config) (*sendAWSKinesisDataFirehose, error) {
	conf := sendAWSKinesisDataFirehoseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_kinesis_data_firehose: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_kinesis_data_firehose"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSKinesisDataFirehose{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:          conf.AWS.Region,
		RoleARN:         conf.AWS.RoleARN,
		MaxRetries:      conf.Retry.Count,
		RetryableErrors: conf.Retry.ErrorMessages,
	})

	agg, err := aggregate.New(aggregate.Config{
		// Firehose limits batch operations to 500 records and 4 MiB.
		Count:    500,
		Size:     sendAWSKinesisDataFirehoseMessageSizeLimit * 4,
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

type sendAWSKinesisDataFirehose struct {
	conf sendAWSKinesisDataFirehoseConfig

	// client is safe for concurrent use.
	client firehose.API

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSKinesisDataFirehose) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	if len(msg.Data()) > sendAWSKinesisDataFirehoseMessageSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSKinesisDataFirehoseRecordSizeLimit)
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
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSKinesisDataFirehose) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSKinesisDataFirehose) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	if _, err := tf.client.PutRecordBatch(ctx, tf.conf.StreamName, data); err != nil {
		return err
	}

	return nil
}
