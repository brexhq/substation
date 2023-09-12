package transform

import (
	"context"
	"encoding/json"
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
const sendAWSSNSMessageSizeLimit = 1024 * 1024 * 256

// errSendAWSSNSMessageSizeLimit is returned when data exceeds the SNS msg
// size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendAWSSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSSNSConfig struct {
	Buffer aggregate.Config `json:"buffer"`
	AWS    iconfig.AWS      `json:"aws"`
	Retry  iconfig.Retry    `json:"retry"`

	// Topic is the AWS SNS topic that data is sent to.
	Topic string `json:"topic"`
}

func (c *sendAWSSNSConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSSNSConfig) Validate() error {
	if c.Topic == "" {
		return fmt.Errorf("topic: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSSNS(_ context.Context, cfg config.Config) (*sendAWSSNS, error) {
	conf := sendAWSSNSConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_sns: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_send_aws_sns: %v", err)
	}

	tf := sendAWSSNS{
		conf: conf,
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:     conf.AWS.Region,
		AssumeRole: conf.AWS.AssumeRole,
		MaxRetries: conf.Retry.Count,
	})

	agg, err := aggregate.New(aggregate.Config{
		// SNS limits batch operations to 10 msgs.
		Count: 10,
		// SNS limits batch operations to 256 KB.
		Size:     sendAWSSNSMessageSizeLimit,
		Interval: conf.Buffer.Interval,
	})
	if err != nil {
		return nil, err
	}

	tf.buffer = agg

	return &tf, nil
}

type sendAWSSNS struct {
	conf sendAWSSNSConfig

	// client is safe for concurrent use.
	client sns.API
	// buffer is safe for concurrent use.
	buffer    *aggregate.Aggregate
	bufferKey string
}

func (tf *sendAWSSNS) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if tf.buffer.Count(tf.bufferKey) == 0 {
			return []*message.Message{msg}, nil
		}

		items := tf.buffer.Get(tf.bufferKey)
		_, err := tf.client.PublishBatch(ctx, tf.conf.Topic, items)
		if err != nil {
			return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
		}

		tf.buffer.Reset(tf.bufferKey)
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSSNSMessageSizeLimit {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", errSendAWSSNSMessageSizeLimit)
	}

	// Send data to SNS only when the buffer is full.
	if ok := tf.buffer.Add(tf.bufferKey, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	items := tf.buffer.Get(tf.bufferKey)
	if _, err := tf.client.PublishBatch(ctx, tf.conf.Topic, items); err != nil {
		return nil, fmt.Errorf("transform: send_aws_sns: %v", err)
	}

	// Reset the buffer and add the msg data.
	tf.buffer.Reset(tf.bufferKey)
	_ = tf.buffer.Add(tf.bufferKey, msg.Data())

	return []*message.Message{msg}, nil
}

func (tf *sendAWSSNS) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
