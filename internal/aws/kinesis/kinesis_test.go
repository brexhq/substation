package kinesis

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

type mockedReceiveMsgs struct {
	kinesisiface.KinesisAPI
	Resp kinesis.PutRecordOutput
}

func (m mockedReceiveMsgs) PutRecordWithContext(ctx aws.Context, in *kinesis.PutRecordInput, opts ...request.Option) (*kinesis.PutRecordOutput, error) {
	return &m.Resp, nil
}

func TestPutRecord(t *testing.T) {
	var tests = []struct {
		resp     kinesis.PutRecordOutput
		expected string
	}{
		{
			resp: kinesis.PutRecordOutput{
				EncryptionType: aws.String("NONE"),
				SequenceNumber: aws.String("ABCDEF"),
				ShardId:        aws.String("XYZ"),
			},
			expected: "ABCDEF",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedReceiveMsgs{Resp: test.resp},
		}
		resp, err := a.PutRecord(ctx, []byte(""), "", "")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if *resp.SequenceNumber != test.expected {
			t.Logf("expected %+v, got %s", resp.SequenceNumber, test.expected)
			t.Fail()
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
	var tests = []struct {
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

// tests that the calculated record size matches the size of returned data
func TestSize(t *testing.T) {
	var tests = []struct {
		data   []byte
		repeat int
		pk     string
	}{
		{
			[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
			1,
			"8Ex8TUWD3dWUMh6dUKaT",
		},
		{
			[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
			235,
			"8Ex8TUWD3dWUMh6dUKaT",
		},
		{
			[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit."),
			5678,
			"8Ex8TUWD3dWUMh6dUKaT",
		},
	}

	rec := Aggregate{}
	rec.New()

	for _, test := range tests {
		b := bytes.Repeat(test.data, test.repeat)
		check := rec.calculateRecordSize(b, test.pk)
		rec.Add(b, test.pk)

		data := rec.Get()
		if check != len(data) {
			t.Logf("expected %v, got %v", len(data), check)
			t.Fail()
		}
	}
}
