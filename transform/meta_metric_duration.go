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
	ID     string         `json:"id"`
	Metric iconfig.Metric `json:"metric"`

	// Transform that has its duration measured.
	//
	// This is deprecated and will be removed in a future release.
	Transform config.Config `json:"transform"`
	// Transforms that have their total duration measured.
	Transforms []config.Config `json:"transforms"`
}

func (c *metaMetricDurationConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newMetaMetricsDuration(ctx context.Context, cfg config.Config) (*metaMetricDuration, error) {
	// conf gets validated when calling metrics.New.
	conf := metaMetricDurationConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform meta_metric_duration: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "meta_metric_duration"
	}

	m, err := metrics.New(ctx, conf.Metric.Destination)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := metaMetricDuration{
		conf:   conf,
		metric: m,
	}

	if conf.Transform.Type != "" {
		tfer, err := New(ctx, conf.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}
		tf.tf = tfer
	}

	tf.tfs = make([]Transformer, len(conf.Transforms))
	for i, t := range conf.Transforms {
		tfer, err := New(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
		}
		tf.tfs[i] = tfer
	}

	return &tf, nil
}

type metaMetricDuration struct {
	conf metaMetricDurationConfig

	tf  Transformer
	tfs []Transformer

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
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		var msgs []*message.Message
		var err error

		if len(tf.tfs) > 0 {
			msgs, err = Apply(ctx, tf.tfs, msg)
		} else {
			msgs, err = tf.tf.Transform(ctx, msg)
		}

		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return msgs, nil
	}

	start := time.Now()
	defer func() {
		tf.duration += time.Since(start)
	}()

	if len(tf.tfs) > 0 {
		return Apply(ctx, tf.tfs, msg)
	}

	return tf.tf.Transform(ctx, msg)
}

func (tf *metaMetricDuration) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
