package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/brexhq/substation"
)

func main() {
	lambda.Start(handler)
}

type validationEvent struct {
	Content string `json:"content"`
	URI     string `json:"uri"`
}

func handler(ctx context.Context, event json.RawMessage) error {
	var e validationEvent
	err := json.Unmarshal(event, &e)
	if err != nil {
		return fmt.Errorf("validation: json: %v (%q)", err, string(event))
	}

	conf, err := base64.StdEncoding.DecodeString(e.Content)
	if err != nil {
		return fmt.Errorf("validation: base64: %v (%q)", err, e.Content)
	}

	cfg := substation.Config{}
	if err := json.Unmarshal(conf, &cfg); err != nil {
		return fmt.Errorf("validation: json: %v (%q)", err, string(conf))
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("validation: substation: %v", err)
	}
	defer sub.Close(ctx)

	return nil
}
