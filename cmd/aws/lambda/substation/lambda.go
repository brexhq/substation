package main

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/message"
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

	// Data and ctrl messages are sent as a group.
	msg := []*message.Message{
		message.New().SetData(evt),
		message.New().AsControl(),
	}

	res, err := sub.Transform(ctx, msg...)
	if err != nil {
		return nil, err
	}

	// Convert transformed messages to a JSON array.
	var output []json.RawMessage
	for _, msg := range res {
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
