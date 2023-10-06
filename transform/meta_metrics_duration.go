package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/metrics"
	"github.com/brexhq/substation/message"
)

type metaMetricsDurationConfig struct {
	Name        string            `json:"name"`
	Attributes  map[string]string `json:"attributes"`
	Destination config.Config     `json:"destination"`
	Transform   config.Config     `json:"transform"`
}

func (c *metaMetricsDurationConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newMetaMetricsDuration(ctx context.Context, cfg config.Config) (*metaMetricsDuration, error) {
	// conf gets validated when calling metrics.New.
	conf := metaMetricsDurationConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_metrics_duration: %v", err)
	}

	m, err := metrics.New(ctx, conf.Destination)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_metrics_duration: %v", err)
	}

	tf := metaMetricsDuration{
		conf:   conf,
		metric: m,
	}

	tfConf, err := json.Marshal(conf.Transform)
	if err != nil {
		return nil, err
	}

	var tfCfg config.Config
	if err := json.Unmarshal(tfConf, &tfCfg); err != nil {
		return nil, err
	}

	tfer, err := New(ctx, tfCfg)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_metrics_duration: %v", err)
	}
	tf.tf = tfer

	return &tf, nil
}

type metaMetricsDuration struct {
	conf metaMetricsDurationConfig

	tf Transformer

	// This is measured in nanoseconds.
	metric   metrics.Generator
	duration time.Duration
}

func (tf *metaMetricsDuration) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Name,
			Value:      tf.duration,
			Attributes: tf.conf.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform: meta_metrics_duration: %v", err)
		}

		return []*message.Message{msg}, nil
	}

	start := time.Now()
	defer func() {
		tf.duration += time.Since(start)
	}()

	return tf.tf.Transform(ctx, msg)
}

func (tf *metaMetricsDuration) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
