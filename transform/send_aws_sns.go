package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sns"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// Records greater than 256 KB in size cannot be
// put into an SNS topic
const sendSNSMessageSizeLimit = 1024 * 1024 * 256

// errSendSNSMessageSizeLimit is returned when data exceeds the SNS message
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSNSConfig struct {
	Buffer  aggregate.Config      `json:"buffer"`
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
	buffer *aggregate.Aggregate
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

	agg, err := aggregate.New(
		aggregate.Config{
			// SNS limits batch operations to 10 messages.
			Count: 10,
			// SNS limits batch operations to 256 KB.
			Size:     sendSNSMessageSizeLimit,
			Interval: conf.Buffer.Interval,
		})
	if err != nil {
		return nil, err
	}

	send.buffer = agg

	return &send, nil
}

func (*sendAWSSNS) Close(context.Context) error {
	return nil
}

func (send *sendAWSSNS) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		if send.buffer.Count() == 0 {
			return []*mess.Message{message}, nil
		}

		items := send.buffer.Get()
		_, err := send.client.PublishBatch(ctx, send.conf.Topic, items)
		if err != nil {
			return nil, fmt.Errorf("send: aws_sns: %v", err)
		}

		send.buffer.Reset()
		return []*mess.Message{message}, nil
	}

	if len(message.Data()) > sendSNSMessageSizeLimit {
		return nil, fmt.Errorf("send: aws_sns: %v", errSendSNSMessageSizeLimit)
	}

	// Send data to SNS only when the buffer is full.
	if ok := send.buffer.Add(message.Data()); ok {
		return []*mess.Message{message}, nil
	}

	items := send.buffer.Get()
	if _, err := send.client.PublishBatch(ctx, send.conf.Topic, items); err != nil {
		return nil, fmt.Errorf("send: aws_sns: %v", err)
	}

	// Reset the buffer and add the message data.
	send.buffer.Reset()
	_ = send.buffer.Add(message.Data())

	return []*mess.Message{message}, nil
}
