package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

func lambdaHandler(ctx context.Context, event json.RawMessage) ([]json.RawMessage, error) {
	evt, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("lambda: %v", err)
	}

	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("lambda: %v", err)
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("lambda: %v", err)
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("sqs handler: %v", err)
	}
	defer sub.Close(ctx)

	// Data and control messages are sent to the transforms as a group.
	var msgs []*message.Message

	data, err := message.New(
		message.SetData(evt),
	)
	if err != nil {
		return nil, err
	}

	msgs = append(msgs, data)

	ctrl, err := message.New(
		message.AsControl(),
	)
	if err != nil {
		return nil, err
	}

	msgs = append(msgs, ctrl)

	msgs, err = transform.Apply(ctx, sub.Transforms(), msgs...)
	if err != nil {
		return nil, err
	}

	var output []json.RawMessage
	for _, msg := range msgs {
		if msg.IsControl() {
			continue
		}

		var rm json.RawMessage
		if err := json.Unmarshal(msg.Data(), &rm); err != nil {
			return nil, fmt.Errorf("lambda sync: %v", err)
		}

		output = append(output, rm)
	}

	return output, nil
}

// lambdaAsyncHandler is triggered by an asynchronous invocation of the Lambda. Read
// more about synchronous invocation here:
// https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html.
//
// This implementation of Substation only supports the object handing pattern.
func lambdaAsyncHandler(ctx context.Context, event json.RawMessage) error {
	if _, err := lambdaHandler(ctx, event); err != nil {
		return err
	}

	return nil
}

// lambdaSyncHandler implements a request-reply service that is triggered by synchronous
// invocation of the Lambda. Read more about synchronous invocation here:
// https://docs.aws.amazon.com/lambda/latest/dg/invocation-sync.html.
//
// This implementation of Substation only supports the object handing pattern.
func lambdaSyncHandler(ctx context.Context, event json.RawMessage) ([]json.RawMessage, error) {
	return lambdaHandler(ctx, event)
}
