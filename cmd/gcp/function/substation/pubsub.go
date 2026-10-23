package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/brexhq/substation/v2/message"
)

// MessagePublishedData is the event data for a Pub/Sub message published event.
type MessagePublishedData struct {
	Message      PubSubMessage `json:"message"`
	Subscription string        `json:"subscription"`
}

// PubSubMessage is a Pub/Sub message delivered through Eventarc.
type PubSubMessage struct {
	Data        []byte            `json:"data"`
	MessageID   string            `json:"messageId"`
	PublishTime time.Time         `json:"publishTime"`
	Attributes  map[string]string `json:"attributes"`
}

type pubsubMetadata struct {
	Subscription string            `json:"subscription"`
	MessageID    string            `json:"messageId"`
	Attributes   map[string]string `json:"attributes"`
}

func pubsubHandler(ctx context.Context, e cloudevents.Event) error {
	cfg, sub, err := loadRuntime(ctx)
	if err != nil {
		return err
	}

	return runAsyncTransform(ctx, sub, cfg.Concurrency, func(ctx context.Context, send messageSender) error {
		var evt MessagePublishedData
		if err := json.Unmarshal(e.Data(), &evt); err != nil {
			return fmt.Errorf("failed to unmarshal event data: %v", err)
		}

		m := pubsubMetadata{
			Subscription: evt.Subscription,
			MessageID:    evt.Message.MessageID,
			Attributes:   evt.Message.Attributes,
		}

		metadata, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("pubsub handler: %v", err)
		}

		send(message.New().SetData(evt.Message.Data).SetMetadata(metadata))

		return nil
	})
}
