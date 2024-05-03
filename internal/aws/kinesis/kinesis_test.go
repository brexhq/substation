package kinesis

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

type mockedPutRecords struct {
	kinesisiface.KinesisAPI
	Resp kinesis.PutRecordsOutput
}

func (m mockedPutRecords) PutRecordsWithContext(ctx aws.Context, in *kinesis.PutRecordsInput, opts ...request.Option) (*kinesis.PutRecordsOutput, error) {
	return &m.Resp, nil
}

func TestPutRecords(t *testing.T) {
	tests := []struct {
		resp     kinesis.PutRecordsOutput
		expected string
	}{
		{
			resp: kinesis.PutRecordsOutput{
				EncryptionType: aws.String("NONE"),
				Records: []*kinesis.PutRecordsResultEntry{
					{
						ErrorCode:      aws.String(""),
						ErrorMessage:   aws.String(""),
						SequenceNumber: aws.String("ABCDEF"),
						ShardId:        aws.String("XYZ"),
					},
				},
			},
			expected: "ABCDEF",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedPutRecords{Resp: test.resp},
		}

		b := [][]byte{
			[]byte(""),
		}
		resp, err := a.PutRecords(ctx, "", "", b)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if *resp.Records[0].SequenceNumber != test.expected {
			t.Errorf("expected %+v, got %s", test.expected, *resp.Records[0].SequenceNumber)
		}
	}
}

type mockedGetTags struct {
	kinesisiface.KinesisAPI
	Resp kinesis.ListTagsForStreamOutput
}

func (m mockedGetTags) ListTagsForStreamWithContext(ctx aws.Context, in *kinesis.ListTagsForStreamInput, opts ...request.Option) (*kinesis.ListTagsForStreamOutput, error) {
	return &m.Resp, nil
}

func TestGetTags(t *testing.T) {
	tests := []struct {
		resp     kinesis.ListTagsForStreamOutput
		expected []*kinesis.Tag
	}{
		{
			resp: kinesis.ListTagsForStreamOutput{
				Tags: []*kinesis.Tag{
					{
						Key:   aws.String("foo"),
						Value: aws.String("bar"),
					},
					{
						Key:   aws.String("baz"),
						Value: aws.String("qux"),
					},
				},
				// can't test recursion via this style of mock
				HasMoreTags: aws.Bool(false),
			},
			expected: []*kinesis.Tag{
				{
					Key:   aws.String("foo"),
					Value: aws.String("bar"),
				},
				{
					Key:   aws.String("baz"),
					Value: aws.String("qux"),
				},
			},
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedGetTags{Resp: test.resp},
		}
		tags, err := a.GetTags(ctx, "")
		if err != nil {
			t.Fatalf("%v", err)
		}

		for idx, test := range test.expected {
			tag := tags[idx]
			if *tag.Key != *test.Key {
				t.Logf("expected %s, got %s", *test.Key, *tag.Key)
				t.Fail()
			}

			if *tag.Value != *test.Value {
				t.Logf("expected %s, got %s", *test.Value, *tag.Value)
				t.Fail()
			}
		}
	}
}
