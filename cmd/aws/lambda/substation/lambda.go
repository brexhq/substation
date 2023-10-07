package main

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

func lambdaHandler(ctx context.Context, event json.RawMessage) ([]json.RawMessage, error) {
	evt, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return nil, err
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Data and control messages are sent to the transforms as a group.
	var msgs []*message.Message

	msg := message.New().SetData(evt)
	msgs = append(msgs, msg)

	ctrl := message.New(message.AsControl())
	msgs = append(msgs, ctrl)

	msgs, err = transform.Apply(ctx, sub.Transforms(), msgs...)
	if err != nil {
		return nil, err
	}

	// Convert transformed Messages to a JSON array.
	var output []json.RawMessage
	for _, msg := range msgs {
		if msg.IsControl() {
			continue
		}

		if !json.Valid(msg.Data()) {
			return nil, errLambdaInvalidJSON
		}

		var rm json.RawMessage
		if err := json.Unmarshal(msg.Data(), &rm); err != nil {
			return nil, err
		}

		output = append(output, rm)
	}

	return output, nil
}
