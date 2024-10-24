package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/metrics"
)

type metaMetricDurationConfig struct {
	// Transforms that have their total duration measured.
	Transforms []config.Config `json:"transforms"`

	ID     string         `json:"id"`
	Metric iconfig.Metric `json:"metric"`
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

	tfs []Transformer

	// This is measured in nanoseconds.
	metric   metrics.Generator
	duration time.Duration
}

func (tf *metaMetricDuration) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.HasFlag(message.IsControl) {
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Metric.Name,
			Value:      tf.duration,
			Attributes: tf.conf.Metric.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		msgs, err := Apply(ctx, tf.tfs, msg)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		return msgs, nil
	}

	start := time.Now()
	defer func() {
		tf.duration += time.Since(start)
	}()

	return Apply(ctx, tf.tfs, msg)
}

func (tf *metaMetricDuration) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
