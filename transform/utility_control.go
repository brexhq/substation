package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type utilityControlConfig struct {
	ID    string        `json:"id"`
	Batch iconfig.Batch `json:"batch"`
}

func (c *utilityControlConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityControl(_ context.Context, cfg config.Config) (*utilityControl, error) {
	conf := utilityControlConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform utility_control: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_control"
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := utilityControl{
		conf: conf,
		agg:  *agg,
	}

	return &tf, nil
}

type utilityControl struct {
	conf utilityControlConfig

	mu  sync.Mutex
	agg aggregate.Aggregate
}

func (tf *utilityControl) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		// If a control message is received, then the aggregation is reset
		// to prevent sending duplicate control messages.
		tf.agg.ResetAll()

		return []*message.Message{msg}, nil
	}

	if ok := tf.agg.Add("", msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	tf.agg.Reset("")
	if ok := tf.agg.Add("", msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}

	ctrl := message.New().AsControl()
	return []*message.Message{ctrl, msg}, nil
}

func (tf *utilityControl) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
