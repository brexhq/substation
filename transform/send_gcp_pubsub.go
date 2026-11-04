package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/pubsub"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// pubsubPublisher publishes a batch of records to a Pub/Sub topic. It is an
// interface so tests can substitute a fake and avoid any network access.
type pubsubPublisher interface {
	Publish(ctx context.Context, orderingKey string, data [][]byte) error
}

type sendGCPPubSubConfig struct {
	// OrderingKey is the message key used to derive an ordering key value.
	//
	// When set, message ordering is enabled on the publisher, records are
	// batched by the extracted value, and each published message is assigned
	// that value as its ordering key. Ordering keys cap throughput at 1 MB/s
	// per key, so ordering is disabled by default (empty).
	OrderingKey string `json:"ordering_key"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	GCP    iconfig.GCP    `json:"gcp"`
}

func (c *sendGCPPubSubConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendGCPPubSubConfig) Validate() error {
	if c.GCP.Resource == "" {
		return fmt.Errorf("gcp.resource: %v", iconfig.ErrMissingRequiredOption)
	}

	if _, _, err := parseGCPPubSubResource(c.GCP.Resource); err != nil {
		return fmt.Errorf("gcp.resource: %v", iconfig.ErrInvalidOption)
	}

	return nil
}

// parseGCPPubSubResource extracts the project and topic from a Pub/Sub topic
// resource name in the form projects/{project}/topics/{topic}.
func parseGCPPubSubResource(resource string) (project, topic string, err error) {
	parts := strings.Split(resource, "/")
	if len(parts) != 4 || parts[0] != "projects" || parts[2] != "topics" || parts[1] == "" || parts[3] == "" {
		return "", "", fmt.Errorf("resource must be in the form projects/{project}/topics/{topic}")
	}

	return parts[1], parts[3], nil
}

func newSendGCPPubSub(ctx context.Context, cfg config.Config) (*sendGCPPubSub, error) {
	conf := sendGCPPubSubConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_gcp_pubsub: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_gcp_pubsub"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	project, topicID, err := parseGCPPubSubResource(conf.GCP.Resource)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendGCPPubSub{
		conf: conf,
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	// Initialize the Pub/Sub client. The client honors PUBSUB_EMULATOR_HOST
	// when it is set.
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	topic := client.Topic(topicID)
	// Ordering defaults off; only enable it when an ordering key is configured.
	topic.EnableMessageOrdering = conf.OrderingKey != ""

	tf.client = client
	tf.publisher = &gcpPubSubTopicPublisher{topic: topic}

	return &tf, nil
}

type sendGCPPubSub struct {
	conf      sendGCPPubSubConfig
	client    *pubsub.Client
	publisher pubsubPublisher

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendGCPPubSub) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	// When an ordering key is configured, batch by its value so every record in
	// a batch shares the same ordering key. Otherwise fall back to the batch key
	// (or a single batch when neither is set).
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if tf.conf.OrderingKey != "" {
		key = msg.GetValue(tf.conf.OrderingKey).String()
	}

	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If data cannot be added after reset, then the batch is misconfigured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendGCPPubSub) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendGCPPubSub) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	// Ordering is off unless an ordering key is configured; when off, no
	// ordering key is passed to the publisher.
	orderingKey := ""
	if tf.conf.OrderingKey != "" {
		orderingKey = key
	}

	ctx = context.WithoutCancel(ctx)
	return tf.publisher.Publish(ctx, orderingKey, data)
}

// gcpPubSubTopicPublisher publishes records to a Pub/Sub topic. Records are
// published concurrently and then confirmed, mirroring the batch-send idiom of
// the AWS senders.
type gcpPubSubTopicPublisher struct {
	topic *pubsub.Topic
}

func (p *gcpPubSubTopicPublisher) Publish(ctx context.Context, orderingKey string, data [][]byte) error {
	results := make([]*pubsub.PublishResult, 0, len(data))
	for _, d := range data {
		m := &pubsub.Message{Data: d}
		if orderingKey != "" {
			m.OrderingKey = orderingKey
		}

		results = append(results, p.topic.Publish(ctx, m))
	}

	// Wait for every result before returning so no records are still in
	// flight when a failed batch is retried.
	var firstErr error
	for _, r := range results {
		if _, err := r.Get(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		// A failed publish pauses all further publishes for the ordering key
		// until it is resumed.
		if orderingKey != "" {
			p.topic.ResumePublish(orderingKey)
		}

		return firstErr
	}

	return nil
}
