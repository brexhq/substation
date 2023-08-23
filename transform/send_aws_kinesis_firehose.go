package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/firehose"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// Records greater than 1000 KiB in size cannot be put into Kinesis Firehose.
const sendKinesisFirehoseMessageSizeLimit = 1024 * 1000

// errSendFirehoseRecordSizeLimit is returned when data exceeds the
// Kinesis Firehose record size limit. If this error occurs,
// then drop or reduce the size of the data before attempting to
// send it to Kinesis Firehose.
var errSendFirehoseRecordSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSKinesisFirehoseConfig struct {
	Buffer  aggregate.Config      `json:"buffer"`
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
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
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Stream == "" {
		return nil, fmt.Errorf("send: aws_sqs: queue: %v", errors.ErrMissingRequiredOption)
	}

	send := sendAWSKinesisFirehose{
		conf: conf,
	}

	agg, err := aggregate.New(
		aggregate.Config{
			// Firehose limits batch operations to 500 records.
			Count: 500,
			// Firehose limits batch operations to 4 MiB.
			Size:     sendKinesisFirehoseMessageSizeLimit * 4,
			Interval: conf.Buffer.Interval,
		})
	if err != nil {
		return nil, err
	}

	send.buffer = agg

	// Setup the AWS client.
	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	return &send, nil
}

func (*sendAWSKinesisFirehose) Close(_ context.Context) error {
	return nil
}

func (send *sendAWSKinesisFirehose) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		if send.buffer.Count() == 0 {
			return []*mess.Message{message}, nil
		}

		items := send.buffer.Get()
		if _, err := send.client.PutRecordBatch(ctx, send.conf.Stream, items); err != nil {
			return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", err)
		}

		send.buffer.Reset()
		return []*mess.Message{message}, nil
	}

	if len(message.Data()) > sendKinesisFirehoseMessageSizeLimit {
		return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", errSendFirehoseRecordSizeLimit)
	}

	// Send data to Kinesis Firehose only when the buffer is full.
	if ok := send.buffer.Add(message.Data()); ok {
		return []*mess.Message{message}, nil
	}

	items := send.buffer.Get()
	if _, err := send.client.PutRecordBatch(ctx, send.conf.Stream, items); err != nil {
		return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer.Reset()
	_ = send.buffer.Add(message.Data())

	return []*mess.Message{message}, nil
}
