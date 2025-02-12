package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// Records greater than 256 KB in size cannot be
// put into an EventBridge bus.
const sendAWSEventBridgeMessageSizeLimit = 1024 * 1024 * 256

// errSendAWSEventBridgeMessageSizeLimit is returned when data
// exceeds  the EventBridge message size limit. If this error
// occurs, then conditions or transforms should be applied to
// either drop or reduce the size of the data.
var errSendAWSEventBridgeMessageSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSEventBridgeConfig struct {
	// Describes the type of the messages sent to EventBridge.
	Description string `json:"description"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *sendAWSEventBridgeConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newSendAWSEventBridge(ctx context.Context, cfg config.Config) (*sendAWSEventBridge, error) {
	conf := sendAWSEventBridgeConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_eventbridge: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_eventbridge"
	}

	if conf.Description == "" {
		// The AWS EventBridge service relies on this value for
		// event routing, so any update to the `conf.Description`
		// variable is considered a BREAKING CHANGE.
		conf.Description = "Substation Transform"
	}

	tf := sendAWSEventBridge{
		conf: conf,
	}

	awsCfg, err := iconfig.NewAWS(ctx, conf.AWS)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = eventbridge.NewFromConfig(awsCfg)

	// Setup the batch.
	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, err
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	return &tf, nil
}

type sendAWSEventBridge struct {
	conf sendAWSEventBridgeConfig

	// client is safe for concurrent use.
	client *eventbridge.Client

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSEventBridge) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.HasFlag(message.IsControl) {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendAWSEventBridgeMessageSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSEventBridgeMessageSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
	}
	return []*message.Message{msg}, nil
}

func (tf *sendAWSEventBridge) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSEventBridge) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	entries := make([]types.PutEventsRequestEntry, len(data))
	for i, d := range data {
		// The AWS EventBridge service relies on this value for
		// event routing, so any update to the `source` variable
		// is considered a BREAKING CHANGE.
		source := fmt.Sprintf("substation.%s", tf.conf.ID)
		detail := string(d)

		entry := types.PutEventsRequestEntry{
			Source:     &source,
			Detail:     &detail,
			DetailType: &tf.conf.Description,
		}

		// If empty, this is the default event bus.
		if tf.conf.AWS.ARN != "" {
			entry.EventBusName = &tf.conf.AWS.ARN
		}

		entries[i] = entry
	}

	ctx = context.WithoutCancel(ctx)
	if _, err = tf.client.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: entries,
	}); err != nil {
		return err
	}

	return nil
}
