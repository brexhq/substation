package sqs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type mockedSendMessage struct {
	sqsiface.SQSAPI
	Resp sqs.SendMessageOutput
}

func (m mockedSendMessage) SendMessageWithContext(ctx aws.Context, in *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	return &m.Resp, nil
}

//lint:ignore ST1003 mocks the AWS API call which does not use correct abbrevation syntax (should be GetQueueURLWithContext)
func (m mockedSendMessage) GetQueueUrlWithContext(ctx aws.Context, in *sqs.GetQueueUrlInput, opts ...request.Option) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("foo"),
	}, nil
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		resp     sqs.SendMessageOutput
		expected string
	}{
		{
			resp: sqs.SendMessageOutput{
				MessageId: aws.String("foo"),
			},
			expected: "foo",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedSendMessage{Resp: test.resp},
		}

		resp, err := a.SendMessage(ctx, []byte(""), "")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if *resp.MessageId != test.expected {
			t.Errorf("expected %+v, got %s", test.expected, *resp.MessageId)
		}
	}
}

type mockedSendMessageBatch struct {
	sqsiface.SQSAPI
	Resp sqs.SendMessageBatchOutput
}

func (m mockedSendMessageBatch) SendMessageBatchWithContext(ctx aws.Context, in *sqs.SendMessageBatchInput, opts ...request.Option) (*sqs.SendMessageBatchOutput, error) {
	return &m.Resp, nil
}

//lint:ignore ST1003 mocks the AWS API call which does not use correct abbrevation syntax (should be GetQueueURLWithContext)
func (m mockedSendMessageBatch) GetQueueUrlWithContext(ctx aws.Context, in *sqs.GetQueueUrlInput, opts ...request.Option) (*sqs.GetQueueUrlOutput, error) {
	return &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String("foo"),
	}, nil
}

func TestSendMessageBatch(t *testing.T) {
	tests := []struct {
		resp     sqs.SendMessageBatchOutput
		expected []string
	}{
		{
			resp: sqs.SendMessageBatchOutput{
				Successful: []*sqs.SendMessageBatchResultEntry{
					{
						MessageId: aws.String("foo"),
					},
					{
						MessageId: aws.String("bar"),
					},
					{
						MessageId: aws.String("baz"),
					},
				},
			},
			expected: []string{"foo", "bar", "baz"},
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedSendMessageBatch{Resp: test.resp},
		}

		resp, err := a.SendMessageBatch(ctx, [][]byte{}, "")
		if err != nil {
			t.Fatalf("%v", err)
		}

		for idx, resp := range resp.Successful {
			if *resp.MessageId != test.expected[idx] {
				t.Errorf("expected %+v, got %s", test.expected[idx], *resp.MessageId)
			}
		}
	}
}
