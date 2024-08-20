package firehose

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
)

type mockedPutRecord struct {
	firehoseiface.FirehoseAPI
	Resp firehose.PutRecordOutput
}

func (m mockedPutRecord) PutRecordWithContext(ctx aws.Context, in *firehose.PutRecordInput, opts ...request.Option) (*firehose.PutRecordOutput, error) {
	return &m.Resp, nil
}

func TestPutRecord(t *testing.T) {
	tests := []struct {
		resp     firehose.PutRecordOutput
		expected string
	}{
		{
			resp: firehose.PutRecordOutput{
				Encrypted: aws.Bool(true),
				RecordId:  aws.String("foo"),
			},
			expected: "foo",
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedPutRecord{Resp: test.resp},
		}
		resp, err := a.PutRecord(ctx, []byte{}, "")
		if err != nil {
			t.Fatalf("%v", err)
		}

		if *resp.RecordId != test.expected {
			t.Errorf("expected %+v, got %s", test.expected, *resp.RecordId)
		}
	}
}

type mockedPutRecordBatch struct {
	firehoseiface.FirehoseAPI
	Resp firehose.PutRecordBatchOutput
}

func (m mockedPutRecordBatch) PutRecordBatchWithContext(ctx aws.Context, in *firehose.PutRecordBatchInput, opts ...request.Option) (*firehose.PutRecordBatchOutput, error) {
	return &m.Resp, nil
}

func TestPutRecordBatch(t *testing.T) {
	tests := []struct {
		resp     firehose.PutRecordBatchOutput
		expected []string
	}{
		{
			resp: firehose.PutRecordBatchOutput{
				Encrypted:      aws.Bool(true),
				FailedPutCount: aws.Int64(0),
				RequestResponses: []*firehose.PutRecordBatchResponseEntry{
					{
						RecordId: aws.String("foo"),
					},
					{
						RecordId: aws.String("bar"),
					},
					{
						RecordId: aws.String("baz"),
					},
				},
			},
			expected: []string{"foo", "bar", "baz"},
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		a := API{
			mockedPutRecordBatch{Resp: test.resp},
		}

		resp, err := a.PutRecordBatch(ctx, "", [][]byte{})
		if err != nil {
			t.Fatalf("%v", err)
		}

		for idx, resp := range resp.RequestResponses {
			if *resp.RecordId != test.expected[idx] {
				t.Errorf("expected %+v, got %s", test.expected[idx], *resp.RecordId)
			}
		}
	}
}
