package eventbridge

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	iaws "github.com/brexhq/substation/internal/aws"
)

const (
	// EventBridge requires a source to be set when sending
	// events, so this value is hardcoded in this package.
	//
	// Updates to this value are a BREAKING CHANGE.
	source = "substation"
)

// New returns a configured EventBridge client.
func New(cfg iaws.Config) *eventbridge.EventBridge {
	conf, sess := iaws.New(cfg)

	c := eventbridge.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the EventBridge API interface.
type API struct {
	Client eventbridgeiface.EventBridgeAPI
}

// Setup creates a new EventBridge client.
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

func (a *API) PutEvents(ctx aws.Context, data [][]byte, detailType, eventBusName string) (*eventbridge.PutEventsOutput, error) {
	entries := make([]*eventbridge.PutEventsRequestEntry, len(data))
	for i, d := range data {
		entries[i] = &eventbridge.PutEventsRequestEntry{
			Source:       aws.String(source),
			Detail:       aws.String(string(d)),
			DetailType:   aws.String(detailType),
			EventBusName: aws.String(eventBusName),
		}
	}

	ctx = context.WithoutCancel(ctx)
	return a.Client.PutEventsWithContext(ctx, &eventbridge.PutEventsInput{
		Entries: entries,
	})
}
