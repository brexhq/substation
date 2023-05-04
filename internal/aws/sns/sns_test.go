package sns

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type mockedPublish struct {
	snsiface.SNSAPI
	Resp sns.PublishOutput
}

func (m mockedPublish) PublishWithContext(ctx aws.Context, in *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
	return &m.Resp, nil
}

func TestPublish(t *testing.T) {
	tests := []struct {
		resp     sns.PublishOutput
		expected string
	}{
		{
			resp: sns.PublishOutput{
				MessageId: aws.String("foo"),
			},
			expected: "foo",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedPublish{Resp: test.resp},
		}

		resp, err := a.Publish(ctx, []byte(""), "")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if *resp.MessageId != test.expected {
			t.Errorf("expected %+v, got %s", test.expected, *resp.MessageId)
		}
	}
}

type mockedPublishBatch struct {
	snsiface.SNSAPI
	Resp sns.PublishBatchOutput
}

func (m mockedPublishBatch) PublishBatchWithContext(ctx aws.Context, in *sns.PublishBatchInput, opts ...request.Option) (*sns.PublishBatchOutput, error) {
	return &m.Resp, nil
}

func TestPublishBatch(t *testing.T) {
	tests := []struct {
		resp     sns.PublishBatchOutput
		expected []string
	}{
		{
			resp: sns.PublishBatchOutput{
				Successful: []*sns.PublishBatchResultEntry{
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
			mockedPublishBatch{Resp: test.resp},
		}

		resp, err := a.PublishBatch(ctx, [][]byte{}, "", false)
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
