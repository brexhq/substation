package transform

import (
	"context"
	"encoding/json"
	"fmt"

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
	Buffer iconfig.Buffer `json:"buffer"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`

	// Stream is the Firehose Delivery Stream that data is sent to.
	Stream string `json:"stream"`
}

func (c *sendAWSKinesisDataFirehoseConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSKinesisDataFirehoseConfig) Validate() error {
	if c.Stream == "" {
		return fmt.Errorf("stream: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSKinesisDataFirehose(_ context.Context, cfg config.Config) (*sendAWSKinesisDataFirehose, error) {
	conf := sendAWSKinesisDataFirehoseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_kinesis_data_firehose: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_kinesis_data_firehose: %v", err)
	}

	tf := sendAWSKinesisDataFirehose{
		conf: conf,
	}

	buffer, err := aggregate.New(aggregate.Config{
		// Firehose limits batch operations to 500 records.
		Count: 500,
		// Firehose limits batch operations to 4 MiB.
		Size:     sendAWSKinesisDataFirehoseMessageSizeLimit * 4,
		Interval: conf.Buffer.Interval,
	})
	if err != nil {
		return nil, err
	}

	tf.buffer = buffer
	tf.bufferKey = conf.Buffer.Key

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

type sendAWSKinesisDataFirehose struct {
	conf sendAWSKinesisDataFirehoseConfig

	// client is safe for concurrent use.
	client firehose.API
	// buffer is safe for concurrent use.
	buffer    *aggregate.Aggregate
	bufferKey string
}

func (tf *sendAWSKinesisDataFirehose) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if tf.buffer.Count(tf.bufferKey) == 0 {
			return []*message.Message{msg}, nil
		}

		items := tf.buffer.Get(tf.bufferKey)
		if _, err := tf.client.PutRecordBatch(ctx, tf.conf.Stream, items); err != nil {
			return nil, fmt.Errorf("transform: send_aws_kinesis_data_firehose: %v", err)
		}

		tf.buffer.Reset(tf.bufferKey)
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSKinesisDataFirehoseMessageSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_firehose: %v", errSendAWSKinesisDataFirehoseRecordSizeLimit)
	}

	// Send data to Kinesis Firehose only when the buffer is full.
	if ok := tf.buffer.Add(tf.bufferKey, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	items := tf.buffer.Get(tf.bufferKey)
	if _, err := tf.client.PutRecordBatch(ctx, tf.conf.Stream, items); err != nil {
		return nil, fmt.Errorf("transform: send_aws_kinesis_data_firehose: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(tf.bufferKey)
	_ = tf.buffer.Add(tf.bufferKey, msg.Data())

	return []*message.Message{msg}, nil
}

func (tf *sendAWSKinesisDataFirehose) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendAWSKinesisDataFirehose) Close(_ context.Context) error {
	return nil
}
