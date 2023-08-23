package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sqs"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// records greater than 256 KB in size cannot be
// put into an SQS queue
const sendSQSMessageSizeLimit = 1024 * 1024 * 256

// errSendSQSMessageSizeLimit is returned when data exceeds the SQS message
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSQSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSQSConfig struct {
	Buffer  aggregate.Config      `json:"buffer"`
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
	// Queue is the AWS SQS queue name that data is sent to.
	Queue string `json:"queue"`
}

type sendAWSSQS struct {
	conf sendAWSSQSConfig

	// client is safe for concurrent use.
	client sqs.API
	// buffer is safe for concurrent use.
	buffer *aggregate.Aggregate
}

func newSendAWSSQS(_ context.Context, cfg config.Config) (*sendAWSSQS, error) {
	conf := sendAWSSQSConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Queue == "" {
		return nil, fmt.Errorf("send: aws_sqs: queue: %v", errors.ErrMissingRequiredOption)
	}

	send := sendAWSSQS{
		conf: conf,
	}

	// Setup the AWS client.
	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	agg, err := aggregate.New(
		aggregate.Config{
			// SQS limits batch operations to 10 messages.
			Count: 10,
			// SQS limits batch operations to 256 KB.
			Size:     sendSQSMessageSizeLimit,
			Interval: conf.Buffer.Interval,
		})
	if err != nil {
		return nil, err
	}

	send.buffer = agg

	return &send, nil
}

func (*sendAWSSQS) Close(context.Context) error {
	return nil
}

func (send *sendAWSSQS) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		if send.buffer.Count() == 0 {
			return []*mess.Message{message}, nil
		}

		items := send.buffer.Get()
		if _, err := send.client.SendMessageBatch(ctx, send.conf.Queue, items); err != nil {
			return nil, fmt.Errorf("send: aws_sqs: %v", err)
		}

		send.buffer.Reset()
		return []*mess.Message{message}, nil
	}

	if len(message.Data()) > sendSQSMessageSizeLimit {
		return nil, fmt.Errorf("send: aws_sqs: %v", errSendSQSMessageSizeLimit)
	}

	// Send data to SQS only when the buffer is full.
	if ok := send.buffer.Add(message.Data()); ok {
		return []*mess.Message{message}, nil
	}

	items := send.buffer.Get()
	if _, err := send.client.SendMessageBatch(ctx, send.conf.Queue, items); err != nil {
		return nil, fmt.Errorf("send: aws_sqs: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer.Reset()
	_ = send.buffer.Add(message.Data())

	return []*mess.Message{message}, nil
}
