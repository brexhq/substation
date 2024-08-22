package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/metrics"
	"github.com/brexhq/substation/message"
)

type utilityMetricFreshnessConfig struct {
	Threshold string         `json:"threshold"`
	Metric    iconfig.Metric `json:"metric"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
}

func (c *utilityMetricFreshnessConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *utilityMetricFreshnessConfig) Validate() error {
	if c.Threshold == "" {
		return fmt.Errorf("threshold: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey == "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newUtilityMetricFreshness(ctx context.Context, cfg config.Config) (*utilityMetricFreshness, error) {
	conf := utilityMetricFreshnessConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform utility_metric_freshness: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_metric_freshness"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	m, err := metrics.New(ctx, conf.Metric.Destination)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	dur, err := time.ParseDuration(conf.Threshold)
	if err != nil {
		return nil, fmt.Errorf("transform %s: duration: %v", conf.ID, err)
	}

	tf := utilityMetricFreshness{
		conf:   conf,
		metric: m,
		dur:    dur,
	}

	return &tf, nil
}

type utilityMetricFreshness struct {
	conf   utilityMetricFreshnessConfig
	metric metrics.Generator
	dur    time.Duration

	success uint32
	failure uint32
}

func (tf *utilityMetricFreshness) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// ctrl messages are handled by only one thread, so the map
	// updates below are safe for concurrency.
	if msg.IsControl() {
		tf.conf.Metric.Attributes["FreshnessType"] = "Success"
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Metric.Name,
			Value:      tf.success,
			Attributes: tf.conf.Metric.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		tf.conf.Metric.Attributes["FreshnessType"] = "Failure"
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Metric.Name,
			Value:      tf.failure,
			Attributes: tf.conf.Metric.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		atomic.StoreUint32(&tf.success, 0)
		atomic.StoreUint32(&tf.failure, 0)
		return []*message.Message{msg}, nil
	}

	// This is a time value expected to be in nanoseconds.
	val := msg.GetValue(tf.conf.Object.SourceKey).Int()
	if val == 0 {
		return []*message.Message{msg}, nil
	}

	ts := time.Unix(0, val)
	if time.Since(ts) < tf.dur {
		atomic.AddUint32(&tf.success, 1)
	} else {
		atomic.AddUint32(&tf.failure, 1)
	}

	return []*message.Message{msg}, nil
}

func (tf *utilityMetricFreshness) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
