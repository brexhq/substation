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

type metaMetricDurationConfig struct {
	Metric    iconfig.Metric `json:"metric"`
	Transform config.Config  `json:"transform"`
}

func (c *metaMetricDurationConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newMetaMetricsDuration(ctx context.Context, cfg config.Config) (*metaMetricDuration, error) {
	// conf gets validated when calling metrics.New.
	conf := metaMetricDurationConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: meta_metric_duration: %v", err)
	}

	m, err := metrics.New(ctx, conf.Metric.Destination)
	if err != nil {
		return nil, fmt.Errorf("transform: meta_metric_duration: %v", err)
	}

	tf := metaMetricDuration{
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
		return nil, fmt.Errorf("transform: meta_metric_duration: %v", err)
	}
	tf.tf = tfer

	return &tf, nil
}

type metaMetricDuration struct {
	conf metaMetricDurationConfig

	tf Transformer

	// This is measured in nanoseconds.
	metric   metrics.Generator
	duration time.Duration
}

func (tf *metaMetricDuration) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Metric.Name,
			Value:      tf.duration,
			Attributes: tf.conf.Metric.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform: meta_metric_duration: %v", err)
		}

		msgs, err := tf.tf.Transform(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("transform: meta_metric_duration: %v", err)
		}

		msgs = append(msgs, msg)
		return msgs, nil
	}

	start := time.Now()
	defer func() {
		tf.duration += time.Since(start)
	}()

	return tf.tf.Transform(ctx, msg)
}

func (tf *metaMetricDuration) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
