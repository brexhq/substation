package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/sqs"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
	"github.com/jshlbrd/go-aggregate"
)

// records greater than 256 KB in size cannot be
// put into an SQS queue
const sendSQSMessageSizeLimit = 1024 * 1024 * 256

// errSendSQSMessageSizeLimit is returned when data exceeds the SQS message
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSQSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSQSConfig struct {
	// Queue is the AWS SQS queue name that data is sent to.
	// TODO(v1.0.0): replace with ARN
	Queue string `json:"queue"`
}

type sendAWSSQS struct {
	conf sendAWSSQSConfig

	// client is safe for concurrent use.
	client sqs.API
	// buffer is safe for concurrent use.
	buffer *aggregate.Bytes
}

func newSendAWSSQS(_ context.Context, cfg config.Config) (*sendAWSSQS, error) {
	conf := sendAWSSQSConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Queue == "" {
		return nil, fmt.Errorf("send: aws_sqs: queue: %v", errors.ErrMissingRequiredOption)
	}

	send := sendAWSSQS{
		conf: conf,
	}

	send.client.Setup()

	// SQS limits messages (both individual and batched)
	// at 256 KB. This buffer will not exceed 256 KB or
	// 10 messages.
	send.buffer = &aggregate.Bytes{}
	send.buffer.New(10, sendSQSMessageSizeLimit)

	return &send, nil
}

func (*sendAWSSQS) Close(context.Context) error {
	return nil
}

func (t *sendAWSSQS) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	control := false
	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		if len(message.Data()) > sendSQSMessageSizeLimit {
			return nil, fmt.Errorf("send: aws_sqs: %v", errSendSQSMessageSizeLimit)
		}

		ok := t.buffer.Add(message.Data())
		if !ok {
			items := t.buffer.Get()
			if _, err := t.client.SendMessageBatch(ctx, t.conf.Queue, items); err != nil {
				return nil, fmt.Errorf("send: aws_sqs: %v", err)
			}

			t.buffer.Reset()
			_ = t.buffer.Add(message.Data())
		}
	}

	// If a control message was received, then items are flushed from the buffer.
	if !control {
		return messages, nil
	}

	if t.buffer.Count() > 0 {
		items := t.buffer.Get()
		if _, err := t.client.SendMessageBatch(ctx, t.conf.Queue, items); err != nil {
			return nil, fmt.Errorf("send: aws_sqs: %v", err)
		}

		t.buffer.Reset()
	}

	return messages, nil
}
