package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// TestPubSubHandler verifies that pubsubHandler decodes the Pub/Sub message
// payload, runs the pipeline on it, and attaches the subscription, messageId,
// and attributes as message metadata.
//
// pubsubHandler is asynchronous and discards transformed messages, so we
// observe the pipeline's effect via a config that (1) copies the whole message
// metadata into the record body under "pubsub" (object_copy from "meta @this")
// and (2) writes the record to a temp file (send_file). Reading that file lets
// us assert both the decoded data and the attached metadata.
func TestPubSubHandler(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "captured.json")

	cfg := fmt.Sprintf(`{"transforms":[
		{"type":"object_copy","settings":{"object":{"source_key":"meta @this","target_key":"pubsub"}}},
		{"type":"send_file","settings":{"file_path":{"prefix":%q}}}
	]}`, outFile)
	setConfig(t, cfg)

	payload := []byte(`{"greeting":"hello"}`)
	mpd := MessagePublishedData{
		Message: PubSubMessage{
			Data:       payload, // marshaled to base64 by encoding/json, as Eventarc delivers it
			MessageID:  "msg-123",
			Attributes: map[string]string{"origin": "example"},
		},
		Subscription: "projects/p/subscriptions/s",
	}

	e := cloudevents.NewEvent()
	e.SetID("evt-1")
	e.SetType("google.cloud.pubsub.topic.v1.messagePublished")
	e.SetSource("//pubsub.googleapis.com/projects/p/topics/t")
	if err := e.SetData(cloudevents.ApplicationJSON, mpd); err != nil {
		t.Fatalf("set event data: %v", err)
	}

	if err := pubsubHandler(context.Background(), e); err != nil {
		t.Fatalf("pubsubHandler: %v", err)
	}

	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read captured output: %v", err)
	}

	var got struct {
		Greeting string         `json:"greeting"`
		PubSub   pubsubMetadata `json:"pubsub"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("captured output is not JSON: %v (%q)", err, raw)
	}

	// The pipeline ran on the decoded message.data.
	if got.Greeting != "hello" {
		t.Fatalf("greeting = %q, want the decoded message payload %q", got.Greeting, "hello")
	}

	// Subscription, messageId, and attributes are attached as metadata.
	if got.PubSub.Subscription != mpd.Subscription {
		t.Fatalf("subscription = %q, want %q", got.PubSub.Subscription, mpd.Subscription)
	}

	if got.PubSub.MessageID != mpd.Message.MessageID {
		t.Fatalf("messageId = %q, want %q", got.PubSub.MessageID, mpd.Message.MessageID)
	}

	if got.PubSub.Attributes["origin"] != "example" {
		t.Fatalf("attributes = %v, want origin=example", got.PubSub.Attributes)
	}
}

// TestPubSubHandlerFlush verifies that messages buffered by the pipeline are
// processed by transforms that honor context cancellation when the pipeline is
// flushed at the end of the invocation.
func TestPubSubHandlerFlush(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "captured.json")

	cfg := fmt.Sprintf(`{"transforms":[
		{"type":"aggregate_to_array","settings":{"object":{"target_key":"items"}}},
		{"type":"object_jq","settings":{"filter":"."}},
		{"type":"send_file","settings":{"file_path":{"prefix":%q}}}
	]}`, outFile)
	setConfig(t, cfg)

	e := cloudevents.NewEvent()
	e.SetID("evt-1")
	e.SetType("google.cloud.pubsub.topic.v1.messagePublished")
	e.SetSource("//pubsub.googleapis.com/projects/p/topics/t")
	if err := e.SetData(cloudevents.ApplicationJSON, MessagePublishedData{
		Message: PubSubMessage{Data: []byte(`{"a":1}`)},
	}); err != nil {
		t.Fatalf("set event data: %v", err)
	}

	if err := pubsubHandler(context.Background(), e); err != nil {
		t.Fatalf("pubsubHandler: %v", err)
	}

	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read captured output: %v", err)
	}

	if !strings.Contains(string(raw), `"items"`) {
		t.Fatalf("output = %q, want the flushed aggregate", raw)
	}
}
