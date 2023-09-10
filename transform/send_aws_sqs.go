package transform

import (
	"context"
	"encoding/json"
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
	Buffer iconfig.Buffer `json:"buffer"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`

	// Queue is the AWS SQS queue name that data is sent to.
	Queue string `json:"queue"`
}

func (c *sendAWSSQSConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSSQSConfig) Validate() error {
	if c.Queue == "" {
		return fmt.Errorf("queue: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type sendAWSSQS struct {
	conf sendAWSSQSConfig

	// client is safe for concurrent use.
	client sqs.API
	// buffer is safe for concurrent use.
	buffer    *aggregate.Aggregate
	bufferKey string
}

func newSendAWSSQS(_ context.Context, cfg config.Config) (*sendAWSSQS, error) {
	conf := sendAWSSQSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_sqs: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_sqs: %v", err)
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

	buffer, err := aggregate.New(aggregate.Config{
		// SQS limits batch operations to 10 msgs.
		Count: 10,
		// SQS limits batch operations to 256 KB.
		Size:     sendSQSMessageSizeLimit,
		Interval: conf.Buffer.Interval,
	})
	if err != nil {
		return nil, err
	}

	tf.buffer = buffer
	tf.bufferKey = conf.Buffer.Key

	return &tf, nil
}

func (tf *sendAWSSQS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if tf.buffer.Count(tf.bufferKey) == 0 {
			return []*message.Message{msg}, nil
		}

		items := tf.buffer.Get(tf.bufferKey)
		if _, err := tf.client.SendMessageBatch(ctx, tf.conf.Queue, items); err != nil {
			return nil, fmt.Errorf("transform: send_aws_sqs: %v", err)
		}

		tf.buffer.Reset(tf.bufferKey)
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendSQSMessageSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_sqs: %v", errSendSQSMessageSizeLimit)
	}

	// Send data to SQS only when the buffer is full.
	if ok := tf.buffer.Add(tf.bufferKey, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	items := tf.buffer.Get(tf.bufferKey)
	if _, err := tf.client.SendMessageBatch(ctx, tf.conf.Queue, items); err != nil {
		return nil, fmt.Errorf("transform: send_aws_sqs: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(tf.bufferKey)
	_ = tf.buffer.Add(tf.bufferKey, msg.Data())

	return []*message.Message{msg}, nil
}

func (tf *sendAWSSQS) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendAWSSQS) Close(context.Context) error {
	return nil
}
