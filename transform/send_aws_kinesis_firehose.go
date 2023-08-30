package transform

import (
	"context"
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
const sendKinesisFirehoseMessageSizeLimit = 1024 * 1000

// errSendFirehoseRecordSizeLimit is returned when data exceeds the
// Kinesis Firehose record size limit. If this error occurs,
// then drop or reduce the size of the data before attempting to
// send it to Kinesis Firehose.
var errSendFirehoseRecordSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSKinesisFirehoseConfig struct {
	Buffer aggregate.Config `json:"buffer"`
	AWS    configAWS        `json:"aws"`
	Retry  configRetry      `json:"retry"`

	// Stream is the Kinesis Firehose Delivery Stream that data is sent to.
	Stream string `json:"stream"`
}

type sendAWSKinesisFirehose struct {
	conf sendAWSKinesisFirehoseConfig

	// client is safe for concurrent use.
	client firehose.API
	// buffer is safe for concurrent use.
	buffer *aggregate.Aggregate
}

func newSendAWSKinesisFirehose(_ context.Context, cfg config.Config) (*sendAWSKinesisFirehose, error) {
	conf := sendAWSKinesisFirehoseConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_kinesis_firehose: %v", err)
	}

	// Validate required options.
	if conf.Stream == "" {
		return nil, fmt.Errorf("transform: new_send_aws_kinesis_firehose: stream: %v", errors.ErrMissingRequiredOption)
	}

	tf := sendAWSKinesisFirehose{
		conf: conf,
	}

	agg, err := aggregate.New(
		aggregate.Config{
			// Firehose limits batch operations to 500 records.
			Count: 500,
			// Firehose limits batch operations to 4 MiB.
			Size:     sendKinesisFirehoseMessageSizeLimit * 4,
			Duration: conf.Buffer.Duration,
		})
	if err != nil {
		return nil, err
	}

	tf.buffer = agg

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	return &tf, nil
}

func (*sendAWSKinesisFirehose) Close(_ context.Context) error {
	return nil
}

func (tf *sendAWSKinesisFirehose) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if tf.buffer.Count() == 0 {
			return []*message.Message{msg}, nil
		}

		items := tf.buffer.Get()
		if _, err := tf.client.PutRecordBatch(ctx, tf.conf.Stream, items); err != nil {
			return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", err)
		}

		tf.buffer.Reset()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendKinesisFirehoseMessageSizeLimit {
		return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", errSendFirehoseRecordSizeLimit)
	}

	// Send data to Kinesis Firehose only when the buffer is full.
	if ok := tf.buffer.Add(msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	items := tf.buffer.Get()
	if _, err := tf.client.PutRecordBatch(ctx, tf.conf.Stream, items); err != nil {
		return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset()
	_ = tf.buffer.Add(msg.Data())

	return []*message.Message{msg}, nil
}
