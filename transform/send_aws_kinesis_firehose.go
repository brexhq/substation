package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/firehose"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
	"github.com/jshlbrd/go-aggregate"
)

// Records greater than 1000 KiB in size cannot be put into Kinesis Firehose.
const sendKinesisFirehoseMessageSizeLimit = 1024 * 1000

// errSendFirehoseRecordSizeLimit is returned when data exceeds the
// Kinesis Firehose record size limit. If this error occurs,
// then drop or reduce the size of the data before attempting to
// send it to Kinesis Firehose.
var errSendFirehoseRecordSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSKinesisFirehoseConfig struct {
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
	buffer *aggregate.Bytes
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

	// Firehose limits Batch operations at up to 4 MiB and
	// 500 records per batch. This buffer will not exceed
	// 3.9 MiB or 500 records.
	send.buffer = &aggregate.Bytes{}
	send.buffer.New(500, sendKinesisFirehoseMessageSizeLimit*4*.99)

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

func (t *sendAWSKinesisFirehose) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	control := false
	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		if len(message.Data()) > sendKinesisFirehoseMessageSizeLimit {
			return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", errSendFirehoseRecordSizeLimit)
		}

		ok := t.buffer.Add(message.Data())
		if !ok {
			items := t.buffer.Get()
			if _, err := t.client.PutRecordBatch(ctx, t.conf.Stream, items); err != nil {
				return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", err)
			}

			t.buffer.Reset()
			_ = t.buffer.Add(message.Data())
		}
	}

	// If a control wasn't received, then data stays in the buffer.
	if !control {
		return messages, nil
	}

	// Flush the buffer.
	if t.buffer.Count() > 0 {
		items := t.buffer.Get()
		if _, err := t.client.PutRecordBatch(ctx, t.conf.Stream, items); err != nil {
			return nil, fmt.Errorf("send: aws_kinesis_firehose: %v", err)
		}

		t.buffer.Reset()
	}

	return messages, nil
}
