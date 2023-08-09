package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sns"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
	"github.com/jshlbrd/go-aggregate"
)

// Records greater than 256 KB in size cannot be
// put into an SNS topic
const sendSNSMessageSizeLimit = 1024 * 1024 * 256

// errSendSNSMessageSizeLimit is returned when data exceeds the SNS message
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSNSConfig struct {
	Auth    _config.ConfigAWSAuth `json:"auth"`
	Request _config.ConfigRequest `json:"request"`
	// ARN is the ARN of the AWS SNS topic that data is sent to.
	Topic string `json:"topic"`
}

type sendAWSSNS struct {
	conf sendAWSSNSConfig

	// client is safe for concurrent use.
	client sns.API
	// buffer is safe for concurrent use.
	buffer *aggregate.Bytes
}

func newSendAWSSNS(_ context.Context, cfg config.Config) (*sendAWSSNS, error) {
	conf := sendAWSSNSConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Topic == "" {
		return nil, fmt.Errorf("send: aws_sns: topic: %v", errors.ErrMissingRequiredOption)
	}

	send := sendAWSSNS{
		conf: conf,
	}

	// Setup the AWS client.
	send.client.Setup(aws.Config{
		Region:     conf.Auth.Region,
		AssumeRole: conf.Auth.AssumeRole,
		MaxRetries: conf.Request.MaxRetries,
	})

	// SNS limits messages (both individual and batched)
	// at 256 KB. This buffer will not exceed 256 KB or
	// 10 messages.
	send.buffer = &aggregate.Bytes{}
	send.buffer.New(10, sendSNSMessageSizeLimit)

	return &send, nil
}

func (*sendAWSSNS) Close(context.Context) error {
	return nil
}

func (t *sendAWSSNS) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	control := false

	for _, message := range messages {
		if message.IsControl() {
			control = true
			continue
		}

		if len(message.Data()) > sendSNSMessageSizeLimit {
			return nil, fmt.Errorf("send: aws_sns: %v", errSendSNSMessageSizeLimit)
		}

		ok := t.buffer.Add(message.Data())
		if !ok {
			items := t.buffer.Get()
			if _, err := t.client.PublishBatch(ctx, t.conf.Topic, items); err != nil {
				return nil, fmt.Errorf("send: aws_sns: %v", err)
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
		_, err := t.client.PublishBatch(ctx, t.conf.Topic, items)
		if err != nil {
			return nil, fmt.Errorf("send: aws_sns: %v", err)
		}

		t.buffer.Reset()
	}

	return messages, nil
}
