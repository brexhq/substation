package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/sqs"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// records greater than 256 KB in size cannot be
// put into an SQS queue
const sendSQSMessageSizeLimit = 1024 * 1024 * 256

// errSendSQSMessageSizeLimit is returned when data exceeds the SQS msg
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendSQSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSQSConfig struct {
	Buffer aggregate.Config `json:"buffer"`
	AWS    configAWS        `json:"aws"`
	Retry  configRetry      `json:"retry"`

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
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_sqs: %v", err)
	}

	// Validate required options.
	if conf.Queue == "" {
		return nil, fmt.Errorf("transform: new_send_aws_sqs: queue: %v", errors.ErrMissingRequiredOption)
	}

	tf := sendAWSSQS{
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
			// SQS limits batch operations to 10 msgs.
			Count: 10,
			// SQS limits batch operations to 256 KB.
			Size:     sendSQSMessageSizeLimit,
			Duration: conf.Buffer.Duration,
		})
	if err != nil {
		return nil, err
	}

	tf.buffer = agg

	return &tf, nil
}

func (*sendAWSSQS) Close(context.Context) error {
	return nil
}

func (tf *sendAWSSQS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if tf.buffer.Count() == 0 {
			return []*message.Message{msg}, nil
		}

		items := tf.buffer.Get()
		if _, err := tf.client.SendMessageBatch(ctx, tf.conf.Queue, items); err != nil {
			return nil, fmt.Errorf("send: aws_sqs: %v", err)
		}

		tf.buffer.Reset()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendSQSMessageSizeLimit {
		return nil, fmt.Errorf("send: aws_sqs: %v", errSendSQSMessageSizeLimit)
	}

	// Send data to SQS only when the buffer is full.
	if ok := tf.buffer.Add(msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	items := tf.buffer.Get()
	if _, err := tf.client.SendMessageBatch(ctx, tf.conf.Queue, items); err != nil {
		return nil, fmt.Errorf("send: aws_sqs: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset()
	_ = tf.buffer.Add(msg.Data())

	return []*message.Message{msg}, nil
}
