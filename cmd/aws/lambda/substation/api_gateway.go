package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

var (
	gateway200Response = events.APIGatewayProxyResponse{StatusCode: 200}
	gateway500Response = events.APIGatewayProxyResponse{StatusCode: 500}
)

type gatewayMetadata struct {
	Resource string            `json:"resource"`
	Path     string            `json:"path"`
	Headers  map[string]string `json:"headers"`
}

func gatewayHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return gateway500Response, fmt.Errorf("gateway: %v", err)
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return gateway500Response, fmt.Errorf("gateway: %v", err)
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return gateway500Response, fmt.Errorf("gateway: %v", err)
	}

	// Create Message metadata.
	m := gatewayMetadata{
		Resource: request.Resource,
		Path:     request.Path,
		Headers:  request.Headers,
	}

	metadata, err := json.Marshal(m)
	if err != nil {
		return gateway500Response, fmt.Errorf("gateway: %v", err)
	}

	// Messages are sent to the transforms as a group.
	var msgs []*message.Message

	b := []byte(request.Body)
	msg := message.New().SetData(b).SetMetadata(metadata)
	msgs = append(msgs, msg)

	ctrl := message.New(message.AsControl())
	msgs = append(msgs, ctrl)

	// Send messages through the transforms.
	if _, err := transform.Apply(ctx, sub.Transforms(), msgs...); err != nil {
		return gateway500Response, fmt.Errorf("gateway: %v", err)
	}

	return gateway200Response, nil
}
