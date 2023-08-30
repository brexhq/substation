package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sns"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// Records greater than 256 KB in size cannot be
// put into an SNS topic
const sendSNSMessageSizeLimit = 1024 * 1024 * 256

// errSendSNSMessageSizeLimit is returned when data exceeds the SNS msg
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSNSConfig struct {
	Buffer aggregate.Config `json:"buffer"`
	AWS    configAWS        `json:"aws"`
	Retry  configRetry      `json:"retry"`

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
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_sns: %v", err)
	}

	// Validate required options.
	if conf.Topic == "" {
		return nil, fmt.Errorf("transform: new_send_aws_sns: topic: %v", errors.ErrMissingRequiredOption)
	}

	tf := sendAWSSNS{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Attempts,
	})

	agg, err := aggregate.New(
		aggregate.Config{
			// SNS limits batch operations to 10 msgs.
			Count: 10,
			// SNS limits batch operations to 256 KB.
			Size:     sendSNSMessageSizeLimit,
			Duration: conf.Buffer.Duration,
		})
	if err != nil {
		return nil, err
	}

	tf.buffer = agg

	return &tf, nil
}

func (*sendAWSSNS) Close(context.Context) error {
	return nil
}

func (tf *sendAWSSNS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if tf.buffer.Count() == 0 {
			return []*message.Message{msg}, nil
		}

		items := tf.buffer.Get()
		_, err := tf.client.PublishBatch(ctx, tf.conf.Topic, items)
		if err != nil {
			return nil, fmt.Errorf("send: aws_sns: %v", err)
		}

		tf.buffer.Reset()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendSNSMessageSizeLimit {
		return nil, fmt.Errorf("send: aws_sns: %v", errSendSNSMessageSizeLimit)
	}

	// Send data to SNS only when the buffer is full.
	if ok := tf.buffer.Add(msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	items := tf.buffer.Get()
	if _, err := tf.client.PublishBatch(ctx, tf.conf.Topic, items); err != nil {
		return nil, fmt.Errorf("send: aws_sns: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset()
	_ = tf.buffer.Add(msg.Data())

	return []*message.Message{msg}, nil
}
