package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
)

var gateway500Response = events.APIGatewayProxyResponse{StatusCode: 500}

type gatewayMetadata struct {
	Resource string            `json:"resource"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
}

func gatewayHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return gateway500Response, err
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return gateway500Response, err
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return gateway500Response, err
	}

	// Create metadata.
	m := gatewayMetadata{
		Resource: request.Resource,
		Path:     request.Path,
		Headers:  request.Headers,
	}

	metadata, err := json.Marshal(m)
	if err != nil {
		return gateway500Response, err
	}

	b := []byte(request.Body)
	msg := []*message.Message{
		message.New().SetData(b).SetMetadata(metadata),
		message.New().AsControl(),
	}

	res, err := sub.Transform(ctx, msg...)
	if err != nil {
		return gateway500Response, err
	}

	// Convert transformed messages to a JSON array.
	var output []json.RawMessage
	for _, msg := range res {
		if msg.IsControl() {
			continue
		}

		if !json.Valid(msg.Data()) {
			return gateway500Response, errLambdaInvalidJSON
		}

		var rm json.RawMessage
		if err := json.Unmarshal(msg.Data(), &rm); err != nil {
			return gateway500Response, err
		}

		output = append(output, rm)
	}

	body, err := json.Marshal(output)
	if err != nil {
		return gateway500Response, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}
