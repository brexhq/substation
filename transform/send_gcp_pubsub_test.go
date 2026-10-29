package transform

import (
	"context"
	"sync"
	"testing"

	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
)

var _ Transformer = &sendGCPPubSub{}

// newTestSendGCPPubSub builds a sendGCPPubSub wired to a fake publisher so
// batching/ordering behavior can be tested without a real Pub/Sub client.
func newTestSendGCPPubSub(conf sendGCPPubSubConfig, pub pubsubPublisher) (*sendGCPPubSub, error) {
	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, err
	}

	return &sendGCPPubSub{
		conf:      conf,
		agg:       agg,
		publisher: pub,
	}, nil
}

type fakePubSubPublisher struct {
	mu    sync.Mutex
	calls []fakePublishCall
}

type fakePublishCall struct {
	orderingKey string
	data        [][]byte
}

func (f *fakePubSubPublisher) Publish(_ context.Context, orderingKey string, data [][]byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	cp := make([][]byte, len(data))
	for i, d := range data {
		b := make([]byte, len(d))
		copy(b, d)
		cp[i] = b
	}

	f.calls = append(f.calls, fakePublishCall{orderingKey: orderingKey, data: cp})
	return nil
}

func (f *fakePubSubPublisher) publishedMessages() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	var n int
	for _, c := range f.calls {
		n += len(c.data)
	}
	return n
}

func TestSendGCPPubSubValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "valid resource",
			settings: map[string]interface{}{"gcp": map[string]interface{}{"resource": "projects/my-project/topics/my-topic"}},
			wantErr:  false,
		},
		{
			name:     "missing resource",
			settings: map[string]interface{}{"gcp": map[string]interface{}{"resource": ""}},
			wantErr:  true,
		},
		{
			name:     "wrong resource format (storage bucket)",
			settings: map[string]interface{}{"gcp": map[string]interface{}{"resource": "projects/_/buckets/my-bucket"}},
			wantErr:  true,
		},
		{
			name:     "wrong resource format (missing topic)",
			settings: map[string]interface{}{"gcp": map[string]interface{}{"resource": "projects/my-project"}},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf := sendGCPPubSubConfig{}
			if err := conf.Decode(test.settings); err != nil {
				t.Fatalf("decode: %v", err)
			}

			err := conf.Validate()
			if test.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !test.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseGCPPubSubResource(t *testing.T) {
	tests := []struct {
		name        string
		resource    string
		wantProject string
		wantTopic   string
		wantErr     bool
	}{
		{"valid", "projects/my-project/topics/my-topic", "my-project", "my-topic", false},
		{"missing topics segment", "projects/my-project", "", "", true},
		{"bucket resource", "projects/_/buckets/my-bucket", "", "", true},
		{"empty project", "projects//topics/my-topic", "", "", true},
		{"empty topic", "projects/my-project/topics/", "", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project, topic, err := parseGCPPubSubResource(test.resource)
			if test.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if project != test.wantProject {
				t.Errorf("project: got %q, want %q", project, test.wantProject)
			}
			if topic != test.wantTopic {
				t.Errorf("topic: got %q, want %q", topic, test.wantTopic)
			}
		})
	}
}

// Ordering keys must default OFF: an unconfigured transform must never set an
// ordering key on published messages.
func TestSendGCPPubSubOrderingKeyDefaultOff(t *testing.T) {
	conf := sendGCPPubSubConfig{}
	if err := conf.Decode(map[string]interface{}{
		"gcp": map[string]interface{}{"resource": "projects/my-project/topics/my-topic"},
	}); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if conf.OrderingKey != "" {
		t.Fatalf("ordering key must default to empty, got %q", conf.OrderingKey)
	}

	fake := &fakePubSubPublisher{}
	tf, err := newTestSendGCPPubSub(conf, fake)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	ctx := context.Background()
	if _, err := tf.Transform(ctx, message.New().SetData([]byte("a"))); err != nil {
		t.Fatalf("transform: %v", err)
	}
	if _, err := tf.Transform(ctx, message.New().AsControl()); err != nil {
		t.Fatalf("transform control: %v", err)
	}

	if got := fake.publishedMessages(); got != 1 {
		t.Fatalf("published messages: got %d, want 1", got)
	}
	for _, c := range fake.calls {
		if c.orderingKey != "" {
			t.Errorf("ordering key must be empty when off, got %q", c.orderingKey)
		}
	}
}

func TestSendGCPPubSubOrderingKeyOn(t *testing.T) {
	conf := sendGCPPubSubConfig{}
	if err := conf.Decode(map[string]interface{}{
		"gcp":          map[string]interface{}{"resource": "projects/my-project/topics/my-topic"},
		"ordering_key": "group",
	}); err != nil {
		t.Fatalf("decode: %v", err)
	}

	fake := &fakePubSubPublisher{}
	tf, err := newTestSendGCPPubSub(conf, fake)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	ctx := context.Background()
	if _, err := tf.Transform(ctx, message.New().SetData([]byte(`{"group":"g1"}`))); err != nil {
		t.Fatalf("transform: %v", err)
	}
	if _, err := tf.Transform(ctx, message.New().AsControl()); err != nil {
		t.Fatalf("transform control: %v", err)
	}

	if len(fake.calls) != 1 {
		t.Fatalf("publish calls: got %d, want 1", len(fake.calls))
	}
	if fake.calls[0].orderingKey != "g1" {
		t.Errorf("ordering key: got %q, want %q", fake.calls[0].orderingKey, "g1")
	}
}

func TestSendGCPPubSubBatchFlushOnControl(t *testing.T) {
	conf := sendGCPPubSubConfig{}
	if err := conf.Decode(map[string]interface{}{
		"gcp":   map[string]interface{}{"resource": "projects/my-project/topics/my-topic"},
		"batch": map[string]interface{}{"count": 10},
	}); err != nil {
		t.Fatalf("decode: %v", err)
	}

	fake := &fakePubSubPublisher{}
	tf, err := newTestSendGCPPubSub(conf, fake)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := tf.Transform(ctx, message.New().SetData([]byte("x"))); err != nil {
			t.Fatalf("transform: %v", err)
		}
	}

	if got := fake.publishedMessages(); got != 0 {
		t.Fatalf("published before flush: got %d, want 0", got)
	}

	if _, err := tf.Transform(ctx, message.New().AsControl()); err != nil {
		t.Fatalf("transform control: %v", err)
	}

	if got := fake.publishedMessages(); got != 3 {
		t.Fatalf("published after flush: got %d, want 3", got)
	}
}
