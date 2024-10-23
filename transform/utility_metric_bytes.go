package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/metrics"
)

type utilityMetricBytesConfig struct {
	Metric iconfig.Metric `json:"metric"`

	ID string `json:"id"`
}

func (c *utilityMetricBytesConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityMetricBytes(ctx context.Context, cfg config.Config) (*utilityMetricBytes, error) {
	// conf gets validated when calling metrics.New.
	conf := utilityMetricBytesConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform utility_metric_bytes: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_metric_bytes"
	}

	m, err := metrics.New(ctx, conf.Metric.Destination)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := utilityMetricBytes{
		conf:   conf,
		metric: m,
	}

	return &tf, nil
}

type utilityMetricBytes struct {
	conf utilityMetricBytesConfig

	metric metrics.Generator
	bytes  uint32
}

func (tf *utilityMetricBytes) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Metric.Name,
			Value:      tf.bytes,
			Attributes: tf.conf.Metric.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		atomic.StoreUint32(&tf.bytes, 0)
		return []*message.Message{msg}, nil
	}

	atomic.AddUint32(&tf.bytes, uint32(len(msg.Data())))
	return []*message.Message{msg}, nil
}

func (tf *utilityMetricBytes) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
