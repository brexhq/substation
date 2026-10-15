package transform

import (
	"context"
	"encoding/json"
	"fmt"

	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type formatOTLPConfig struct {
	// Signal is the OTLP signal in the payload. Must be one of: logs, metrics.
	Signal string `json:"signal"`

	ID string `json:"id"`
}

func (c *formatOTLPConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *formatOTLPConfig) Validate() error {
	switch c.Signal {
	case "logs", "metrics":
		return nil
	case "":
		return fmt.Errorf("signal: %v", iconfig.ErrMissingRequiredOption)
	default:
		return fmt.Errorf("signal %s: %v", c.Signal, iconfig.ErrInvalidOption)
	}
}

func newFormatFromOTLP(_ context.Context, cfg config.Config) (*formatFromOTLP, error) {
	conf := formatOTLPConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_from_otlp: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_from_otlp"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := formatFromOTLP{
		conf: conf,
	}

	return &tf, nil
}

// formatFromOTLP splits an OTLP/HTTP protobuf export request into one JSON message
// per log record or metric data point. Each message is a ResourceLogs or
// ResourceMetrics holding exactly one resource, scope, and record.
//
// Output is protojson with proto field names (time_unix_nano, enum names), not the
// OTLP/JSON wire encoding, which uses lowerCamelCase, integer enums, and hex IDs.
type formatFromOTLP struct {
	conf formatOTLPConfig
}

func (tf *formatFromOTLP) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var records []proto.Message
	var err error

	switch tf.conf.Signal {
	case "logs":
		records, err = splitOTLPLogs(msg.Data())
	case "metrics":
		records, err = splitOTLPMetrics(msg.Data())
	}

	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}

	msgs := make([]*message.Message, 0, len(records))
	for _, r := range records {
		data, err := opts.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		msgs = append(msgs, message.New().SetData(data).SetMetadata(msg.Metadata()))
	}

	return msgs, nil
}

func (tf *formatFromOTLP) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

// unmarshalOTLP rejects unknown fields, because a payload of the wrong signal (or
// not OTLP at all) usually decodes without error into unknown fields. This also
// rejects fields added to OTLP after the pinned otlp module version.
func unmarshalOTLP(data []byte, m proto.Message) error {
	if err := proto.Unmarshal(data, m); err != nil {
		return err
	}

	return protorange.Range(m.ProtoReflect(), func(v protopath.Values) error {
		msg, ok := v.Index(-1).Value.Interface().(protoreflect.Message)
		if ok && len(msg.GetUnknown()) > 0 {
			return fmt.Errorf("unknown fields at %v: payload does not match the configured signal", v.Path)
		}

		return nil
	})
}

func splitOTLPLogs(data []byte) ([]proto.Message, error) {
	// LogsData is wire-compatible with ExportLogsServiceRequest, and avoids the
	// collector package's gRPC gateway dependency.
	req := &logsv1.LogsData{}
	if err := unmarshalOTLP(data, req); err != nil {
		return nil, err
	}

	var out []proto.Message
	for _, rl := range req.GetResourceLogs() {
		for _, sl := range rl.GetScopeLogs() {
			for _, lr := range sl.GetLogRecords() {
				out = append(out, &logsv1.ResourceLogs{
					Resource:  rl.GetResource(),
					SchemaUrl: rl.GetSchemaUrl(),
					ScopeLogs: []*logsv1.ScopeLogs{{
						Scope:      sl.GetScope(),
						SchemaUrl:  sl.GetSchemaUrl(),
						LogRecords: []*logsv1.LogRecord{lr},
					}},
				})
			}
		}
	}

	return out, nil
}

func splitOTLPMetrics(data []byte) ([]proto.Message, error) {
	// MetricsData is wire-compatible with ExportMetricsServiceRequest.
	req := &metricsv1.MetricsData{}
	if err := unmarshalOTLP(data, req); err != nil {
		return nil, err
	}

	var out []proto.Message
	for _, rm := range req.GetResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				for _, single := range splitOTLPMetric(m) {
					out = append(out, &metricsv1.ResourceMetrics{
						Resource:  rm.GetResource(),
						SchemaUrl: rm.GetSchemaUrl(),
						ScopeMetrics: []*metricsv1.ScopeMetrics{{
							Scope:     sm.GetScope(),
							SchemaUrl: sm.GetSchemaUrl(),
							Metrics:   []*metricsv1.Metric{single},
						}},
					})
				}
			}
		}
	}

	return out, nil
}

// splitOTLPMetric returns one copy of the metric per data point.
func splitOTLPMetric(m *metricsv1.Metric) []*metricsv1.Metric {
	header := func() *metricsv1.Metric {
		return &metricsv1.Metric{
			Name:        m.GetName(),
			Description: m.GetDescription(),
			Unit:        m.GetUnit(),
			Metadata:    m.GetMetadata(),
		}
	}

	var out []*metricsv1.Metric
	switch d := m.GetData().(type) {
	case *metricsv1.Metric_Sum:
		for _, dp := range d.Sum.GetDataPoints() {
			c := header()
			c.Data = &metricsv1.Metric_Sum{Sum: &metricsv1.Sum{
				DataPoints:             []*metricsv1.NumberDataPoint{dp},
				AggregationTemporality: d.Sum.GetAggregationTemporality(),
				IsMonotonic:            d.Sum.GetIsMonotonic(),
			}}
			out = append(out, c)
		}
	case *metricsv1.Metric_Gauge:
		for _, dp := range d.Gauge.GetDataPoints() {
			c := header()
			c.Data = &metricsv1.Metric_Gauge{Gauge: &metricsv1.Gauge{
				DataPoints: []*metricsv1.NumberDataPoint{dp},
			}}
			out = append(out, c)
		}
	case *metricsv1.Metric_Histogram:
		for _, dp := range d.Histogram.GetDataPoints() {
			c := header()
			c.Data = &metricsv1.Metric_Histogram{Histogram: &metricsv1.Histogram{
				DataPoints:             []*metricsv1.HistogramDataPoint{dp},
				AggregationTemporality: d.Histogram.GetAggregationTemporality(),
			}}
			out = append(out, c)
		}
	case *metricsv1.Metric_ExponentialHistogram:
		for _, dp := range d.ExponentialHistogram.GetDataPoints() {
			c := header()
			c.Data = &metricsv1.Metric_ExponentialHistogram{ExponentialHistogram: &metricsv1.ExponentialHistogram{
				DataPoints:             []*metricsv1.ExponentialHistogramDataPoint{dp},
				AggregationTemporality: d.ExponentialHistogram.GetAggregationTemporality(),
			}}
			out = append(out, c)
		}
	case *metricsv1.Metric_Summary:
		for _, dp := range d.Summary.GetDataPoints() {
			c := header()
			c.Data = &metricsv1.Metric_Summary{Summary: &metricsv1.Summary{
				DataPoints: []*metricsv1.SummaryDataPoint{dp},
			}}
			out = append(out, c)
		}
	default:
		out = append(out, m)
	}

	// A metric with no data points still carries its name and resource.
	if len(out) == 0 {
		out = append(out, m)
	}

	return out
}
