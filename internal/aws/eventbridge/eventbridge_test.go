package eventbridge

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
)

type mockedPutEvents struct {
	eventbridgeiface.EventBridgeAPI
	Resp eventbridge.PutEventsOutput
}

func (m mockedPutEvents) PutEventsWithContext(ctx aws.Context, input *eventbridge.PutEventsInput, opts ...request.Option) (*eventbridge.PutEventsOutput, error) {
	return &m.Resp, nil
}

func TestPutEvents(t *testing.T) {
	tests := []struct {
		resp     eventbridge.PutEventsOutput
		expected string
	}{
		{
			resp: eventbridge.PutEventsOutput{
				FailedEntryCount: aws.Int64(0),
				Entries: []*eventbridge.PutEventsResultEntry{
					{
						EventId: aws.String("abcdefg"),
					},
				},
			},
			expected: "abcdefg",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedPutEvents{Resp: test.resp},
		}

		resp, err := a.PutEvents(ctx, [][]byte{}, "", "")
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}

		if *resp.FailedEntryCount != 0 {
			t.Errorf("expected %+v, got %d", *resp.FailedEntryCount, 0)
		}

		if *resp.Entries[0].EventId != test.expected {
			t.Errorf("expected %+v, got %s", *resp.Entries[0].EventId, test.expected)
		}
	}
}
