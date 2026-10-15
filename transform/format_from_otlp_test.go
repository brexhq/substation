package transform

import (
	"context"
	"testing"

	"github.com/tidwall/gjson"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &formatFromOTLP{}

func otlpStr(k, v string) *commonv1.KeyValue {
	return &commonv1.KeyValue{Key: k, Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: v}}}
}

func otlpInt(k string, v int64) *commonv1.KeyValue {
	return &commonv1.KeyValue{Key: k, Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: v}}}
}

func otlpNumberPoint(v int64, attrs ...*commonv1.KeyValue) *metricsv1.NumberDataPoint {
	return &metricsv1.NumberDataPoint{Attributes: attrs, Value: &metricsv1.NumberDataPoint_AsInt{AsInt: v}}
}

// Two resources (users), one with two log records, to check that each output
// message keeps its own resource.
func otlpTestLogs(t *testing.T) []byte {
	t.Helper()

	scope := &commonv1.InstrumentationScope{Name: "example.scope", Version: "1.0.0"}
	req := &logsv1.LogsData{ResourceLogs: []*logsv1.ResourceLogs{
		{
			Resource: &resourcev1.Resource{Attributes: []*commonv1.KeyValue{
				otlpStr("service.name", "example"),
				otlpStr("deployment.environment", "prod"),
				otlpStr("user.name", "alice"),
			}},
			ScopeLogs: []*logsv1.ScopeLogs{{
				Scope: scope,
				LogRecords: []*logsv1.LogRecord{
					{
						TimeUnixNano:   1792083019401769000,
						SeverityNumber: logsv1.SeverityNumber_SEVERITY_NUMBER_INFO,
						EventName:      "example.process.exec",
						Body:           &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "process_exec"}},
						Attributes: []*commonv1.KeyValue{
							otlpStr("event.id", "evt-1"),
							otlpStr("process.command_line", "ls -la"),
						},
					},
					{
						TimeUnixNano:   1792083019401770000,
						SeverityNumber: logsv1.SeverityNumber_SEVERITY_NUMBER_WARN,
						EventName:      "example.policy.decision",
						Attributes:     []*commonv1.KeyValue{otlpStr("event.id", "evt-2")},
					},
				},
			}},
		},
		{
			Resource: &resourcev1.Resource{Attributes: []*commonv1.KeyValue{
				otlpStr("service.name", "example"),
				otlpStr("user.name", "bob"),
			}},
			ScopeLogs: []*logsv1.ScopeLogs{{
				Scope:      scope,
				LogRecords: []*logsv1.LogRecord{{EventName: "example.api.request", Attributes: []*commonv1.KeyValue{otlpStr("event.id", "evt-3")}}},
			}},
		},
	}}

	b, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

// One delta sum with two data points (two series) and one gauge with one.
func otlpTestMetrics(t *testing.T) []byte {
	t.Helper()

	req := &metricsv1.MetricsData{ResourceMetrics: []*metricsv1.ResourceMetrics{{
		Resource: &resourcev1.Resource{Attributes: []*commonv1.KeyValue{otlpStr("service.name", "example")}},
		ScopeMetrics: []*metricsv1.ScopeMetrics{{
			Scope: &commonv1.InstrumentationScope{Name: "example.scope"},
			Metrics: []*metricsv1.Metric{
				{
					Name: "example.token.usage",
					Unit: "{token}",
					Data: &metricsv1.Metric_Sum{Sum: &metricsv1.Sum{
						AggregationTemporality: metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
						IsMonotonic:            true,
						DataPoints: []*metricsv1.NumberDataPoint{
							otlpNumberPoint(100, otlpStr("token.type", "input")),
							otlpNumberPoint(40, otlpStr("token.type", "output")),
						},
					}},
				},
				{
					Name: "example.gauge",
					Data: &metricsv1.Metric_Gauge{Gauge: &metricsv1.Gauge{
						DataPoints: []*metricsv1.NumberDataPoint{otlpNumberPoint(7, otlpInt("k", 1))},
					}},
				},
			},
		}},
	}}}

	b, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestFormatFromOTLPLogs(t *testing.T) {
	ctx := context.TODO()

	tf, err := newFormatFromOTLP(ctx, config.Config{Settings: map[string]interface{}{"signal": "logs"}})
	if err != nil {
		t.Fatal(err)
	}

	meta := []byte(`{"partitionKey":"logs"}`)
	msgs, err := tf.Transform(ctx, message.New().SetData(otlpTestLogs(t)).SetMetadata(meta))
	if err != nil {
		t.Fatal(err)
	}

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	expected := []struct {
		user, event, id string
	}{
		{"alice", "example.process.exec", "evt-1"},
		{"alice", "example.policy.decision", "evt-2"},
		{"bob", "example.api.request", "evt-3"},
	}

	for i, m := range msgs {
		d := string(m.Data())
		rec := "scope_logs.0.log_records.0"

		if n := gjson.Get(d, "scope_logs.#").Int(); n != 1 {
			t.Errorf("message %d: expected 1 scope, got %d", i, n)
		}
		if n := gjson.Get(d, "scope_logs.0.log_records.#").Int(); n != 1 {
			t.Errorf("message %d: expected 1 record, got %d", i, n)
		}
		if got := gjson.Get(d, `resource.attributes.#(key=="user.name").value.string_value`).String(); got != expected[i].user {
			t.Errorf("message %d: expected user %s, got %s", i, expected[i].user, got)
		}
		if got := gjson.Get(d, rec+".event_name").String(); got != expected[i].event {
			t.Errorf("message %d: expected event_name %s, got %s", i, expected[i].event, got)
		}
		if got := gjson.Get(d, rec+`.attributes.#(key=="event.id").value.string_value`).String(); got != expected[i].id {
			t.Errorf("message %d: expected event id %s, got %s", i, expected[i].id, got)
		}
		if got := gjson.Get(d, "scope_logs.0.scope.name").String(); got != "example.scope" {
			t.Errorf("message %d: expected scope example.scope, got %s", i, got)
		}
		if string(m.Metadata()) != string(meta) {
			t.Errorf("message %d: metadata not preserved: %s", i, m.Metadata())
		}
	}

	// protojson encodes 64-bit integers as strings, per the OTLP/JSON spec.
	if got := gjson.Get(string(msgs[0].Data()), "scope_logs.0.log_records.0.time_unix_nano").String(); got != "1792083019401769000" {
		t.Errorf("expected time_unix_nano 1792083019401769000, got %s", got)
	}
}

func TestFormatFromOTLPMetrics(t *testing.T) {
	ctx := context.TODO()

	tf, err := newFormatFromOTLP(ctx, config.Config{Settings: map[string]interface{}{"signal": "metrics"}})
	if err != nil {
		t.Fatal(err)
	}

	msgs, err := tf.Transform(ctx, message.New().SetData(otlpTestMetrics(t)))
	if err != nil {
		t.Fatal(err)
	}

	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	expected := []struct {
		name, path string
		value      string
	}{
		{"example.token.usage", "sum.data_points", "100"},
		{"example.token.usage", "sum.data_points", "40"},
		{"example.gauge", "gauge.data_points", "7"},
	}

	for i, m := range msgs {
		d := string(m.Data())
		metric := "scope_metrics.0.metrics.0"

		if got := gjson.Get(d, metric+".name").String(); got != expected[i].name {
			t.Errorf("message %d: expected name %s, got %s", i, expected[i].name, got)
		}
		if n := gjson.Get(d, metric+"."+expected[i].path+".#").Int(); n != 1 {
			t.Errorf("message %d: expected 1 data point, got %d", i, n)
		}
		if got := gjson.Get(d, metric+"."+expected[i].path+".0.as_int").String(); got != expected[i].value {
			t.Errorf("message %d: expected value %s, got %s", i, expected[i].value, got)
		}
	}

	// Sum properties must survive the split, or deltas are misread as cumulative.
	d := string(msgs[0].Data())
	if got := gjson.Get(d, "scope_metrics.0.metrics.0.sum.aggregation_temporality").String(); got != "AGGREGATION_TEMPORALITY_DELTA" {
		t.Errorf("expected delta temporality, got %s", got)
	}
	if !gjson.Get(d, "scope_metrics.0.metrics.0.sum.is_monotonic").Bool() {
		t.Error("expected is_monotonic to be preserved")
	}
}

func TestFormatFromOTLPErrors(t *testing.T) {
	ctx := context.TODO()

	for _, s := range []map[string]interface{}{{}, {"signal": "traces"}} {
		if _, err := newFormatFromOTLP(ctx, config.Config{Settings: s}); err == nil {
			t.Errorf("expected config error for settings %v", s)
		}
	}

	tf, err := newFormatFromOTLP(ctx, config.Config{Settings: map[string]interface{}{"signal": "logs"}})
	if err != nil {
		t.Fatal(err)
	}

	for name, data := range map[string][]byte{
		// Fails protobuf parsing outright.
		"otlp_json": []byte(`{"resourceLogs":[]}`),
		// Parses as protobuf, but only into unknown fields.
		"not_otlp": []byte(`{|`),
		// Parses as logs, with the metrics fields landing as unknown fields.
		"wrong_signal": otlpTestMetrics(t),
	} {
		if _, err := tf.Transform(ctx, message.New().SetData(data)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}

	// An empty export is valid and produces no messages.
	msgs, err := tf.Transform(ctx, message.New().SetData([]byte{}))
	if err != nil || len(msgs) != 0 {
		t.Errorf("expected no messages and no error for an empty export, got %d, %v", len(msgs), err)
	}

	ctrl := message.New().AsControl()
	msgs, err = tf.Transform(ctx, ctrl)
	if err != nil || len(msgs) != 1 || !msgs[0].IsControl() {
		t.Errorf("expected control message to pass through, got %v, %v", msgs, err)
	}
}

// Every metric type keeps its type-level fields through the split, and a metric
// with no data points is still emitted.
func TestFormatFromOTLPMetricTypes(t *testing.T) {
	ctx := context.TODO()

	cumulative := metricsv1.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
	req := &metricsv1.MetricsData{ResourceMetrics: []*metricsv1.ResourceMetrics{{
		Resource:  &resourcev1.Resource{Attributes: []*commonv1.KeyValue{otlpStr("service.name", "example")}},
		SchemaUrl: "https://opentelemetry.io/schemas/1.26.0",
		ScopeMetrics: []*metricsv1.ScopeMetrics{{
			Scope:     &commonv1.InstrumentationScope{Name: "example.scope"},
			SchemaUrl: "https://example.com/scope-schema",
			Metrics: []*metricsv1.Metric{
				{Name: "hist", Unit: "ms", Data: &metricsv1.Metric_Histogram{Histogram: &metricsv1.Histogram{
					AggregationTemporality: cumulative,
					DataPoints:             []*metricsv1.HistogramDataPoint{{Count: 3}, {Count: 5}},
				}}},
				{Name: "exphist", Data: &metricsv1.Metric_ExponentialHistogram{ExponentialHistogram: &metricsv1.ExponentialHistogram{
					AggregationTemporality: cumulative,
					DataPoints:             []*metricsv1.ExponentialHistogramDataPoint{{Count: 2}},
				}}},
				{Name: "summary", Data: &metricsv1.Metric_Summary{Summary: &metricsv1.Summary{
					DataPoints: []*metricsv1.SummaryDataPoint{{Count: 4}, {Count: 6}},
				}}},
				{Name: "empty_sum", Data: &metricsv1.Metric_Sum{Sum: &metricsv1.Sum{}}},
				{Name: "no_data"},
			},
		}},
	}}}

	b, err := proto.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	tf, err := newFormatFromOTLP(ctx, config.Config{Settings: map[string]interface{}{"signal": "metrics"}})
	if err != nil {
		t.Fatal(err)
	}

	msgs, err := tf.Transform(ctx, message.New().SetData(b))
	if err != nil {
		t.Fatal(err)
	}

	expected := []struct {
		name, path, count, temporality string
	}{
		{"hist", "histogram", "3", "AGGREGATION_TEMPORALITY_CUMULATIVE"},
		{"hist", "histogram", "5", "AGGREGATION_TEMPORALITY_CUMULATIVE"},
		{"exphist", "exponential_histogram", "2", "AGGREGATION_TEMPORALITY_CUMULATIVE"},
		{"summary", "summary", "4", ""},
		{"summary", "summary", "6", ""},
		{"empty_sum", "", "", ""},
		{"no_data", "", "", ""},
	}

	if len(msgs) != len(expected) {
		t.Fatalf("expected %d messages, got %d", len(expected), len(msgs))
	}

	for i, m := range msgs {
		d := string(m.Data())
		metric := "scope_metrics.0.metrics.0"
		e := expected[i]

		if got := gjson.Get(d, metric+".name").String(); got != e.name {
			t.Errorf("message %d: expected name %s, got %s", i, e.name, got)
		}
		if got := gjson.Get(d, "schema_url").String(); got != "https://opentelemetry.io/schemas/1.26.0" {
			t.Errorf("message %d: resource schema_url not preserved: %s", i, got)
		}
		if got := gjson.Get(d, "scope_metrics.0.schema_url").String(); got != "https://example.com/scope-schema" {
			t.Errorf("message %d: scope schema_url not preserved: %s", i, got)
		}
		if e.path == "" {
			continue
		}
		if n := gjson.Get(d, metric+"."+e.path+".data_points.#").Int(); n != 1 {
			t.Errorf("message %d: expected 1 data point, got %d", i, n)
		}
		if got := gjson.Get(d, metric+"."+e.path+".data_points.0.count").String(); got != e.count {
			t.Errorf("message %d: expected count %s, got %s", i, e.count, got)
		}
		if got := gjson.Get(d, metric+"."+e.path+".aggregation_temporality").String(); got != e.temporality {
			t.Errorf("message %d: expected temporality %q, got %q", i, e.temporality, got)
		}
	}

	if got := gjson.Get(string(msgs[0].Data()), "scope_metrics.0.metrics.0.unit").String(); got != "ms" {
		t.Errorf("expected unit ms to be preserved, got %s", got)
	}
}

func BenchmarkFormatFromOTLPLogs(b *testing.B) {
	ctx := context.TODO()
	tf, _ := newFormatFromOTLP(ctx, config.Config{Settings: map[string]interface{}{"signal": "logs"}})
	data := otlpTestLogs(&testing.T{})

	for i := 0; i < b.N; i++ {
		_, _ = tf.Transform(ctx, message.New().SetData(data))
	}
}
