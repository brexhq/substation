package kinesis

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	rec "github.com/awslabs/kinesis-aggregation/go/records"
	iaws "github.com/brexhq/substation/internal/aws"

	//nolint: staticcheck // not ready to switch package
	"github.com/golang/protobuf/proto"
)

// Aggregate produces a KPL-compliant Kinesis record
type Aggregate struct {
	Record       *rec.AggregatedRecord
	Count        int
	PartitionKey string
}

// New creates a new Kinesis record with default values
// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L167
func (a *Aggregate) New() {
	a.Record = &rec.AggregatedRecord{}
	a.Count = 0

	a.PartitionKey = ""
	a.Record.PartitionKeyTable = make([]string, 0)
}

// Add inserts a Kinesis record into an aggregated Kinesis record
// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L382
func (a *Aggregate) Add(data []byte, partitionKey string) bool {
	// https://docs.aws.amazon.com/streams/latest/dev/key-concepts.html#partition-key
	if len(partitionKey) > 256 {
		partitionKey = partitionKey[0:256]
	}

	// grab the first parition key in the set of events
	if a.PartitionKey == "" {
		a.PartitionKey = partitionKey
	}

	pki := uint64(a.Count)
	r := &rec.Record{
		PartitionKeyIndex: &pki,
		Data:              data,
	}

	a.Record.Records = append(a.Record.Records, r)
	a.Record.PartitionKeyTable = append(a.Record.PartitionKeyTable, partitionKey)
	a.Count++

	return true
}

// Get returns a KPL-compliant compressed Kinesis record
// https://github.com/awslabs/kinesis-aggregation/blob/398fbd4b430d4bf590431b301d03cbbc94279cef/python/aws_kinesis_agg/aggregator.py#L293
func (a *Aggregate) Get() []byte {
	data, _ := proto.Marshal(a.Record)
	md5Hash := md5.Sum(data)

	record := []byte("\xf3\x89\x9a\xc2")
	record = append(record, data...)
	record = append(record, md5Hash[:]...)

	return record
}

// ConvertEventsRecords converts Kinesis records between the Lambda and Go SDK packages. This is required for deaggregating Kinesis records processed by AWS Lambda.
func ConvertEventsRecords(records []events.KinesisEventRecord) []*kinesis.Record {
	output := make([]*kinesis.Record, 0)

	for _, r := range records {
		// ApproximateArrivalTimestamp is events.SecondsEpochTime which serializes time.Time
		ts := r.Kinesis.ApproximateArrivalTimestamp.UTC()
		output = append(output, &kinesis.Record{
			ApproximateArrivalTimestamp: &ts,
			Data:                        r.Kinesis.Data,
			EncryptionType:              &r.Kinesis.EncryptionType,
			PartitionKey:                &r.Kinesis.PartitionKey,
			SequenceNumber:              &r.Kinesis.SequenceNumber,
		})
	}

	return output
}

// New returns a configured Kinesis client.
func New(cfg iaws.Config) *kinesis.Kinesis {
	conf, sess := iaws.New(cfg)

	c := kinesis.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the Kinesis API interface.
type API struct {
	Client kinesisiface.KinesisAPI
}

// Setup creates a new Kinesis client.
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// PutRecords is a convenience wrapper for putting multiple records into a Kinesis stream.
func (a *API) PutRecords(ctx aws.Context, stream, partitionKey string, data [][]byte) (*kinesis.PutRecordsOutput, error) {
	var records []*kinesis.PutRecordsRequestEntry

	ctx = context.WithoutCancel(ctx)
	for _, d := range data {
		records = append(records, &kinesis.PutRecordsRequestEntry{
			Data:         d,
			PartitionKey: aws.String(partitionKey),
		})
	}

	resp, err := a.Client.PutRecordsWithContext(
		ctx,
		&kinesis.PutRecordsInput{
			Records:    records,
			StreamName: aws.String(stream),
		},
	)

	// If any record fails, then the record is recursively retried.
	if resp.FailedRecordCount != nil && *resp.FailedRecordCount > 0 {
		var retry [][]byte

		for idx, r := range resp.Records {
			if r.ErrorCode != nil {
				retry = append(retry, data[idx])
			}
		}

		if len(retry) > 0 {
			return a.PutRecords(ctx, stream, partitionKey, retry)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("put_records: stream %s: %v", stream, err)
	}

	return resp, nil
}

// ActiveShards returns the number of in-use shards for a Kinesis stream.
func (a *API) ActiveShards(ctx aws.Context, stream string) (int64, error) {
	var shards int64
	params := &kinesis.ListShardsInput{
		StreamName: aws.String(stream),
	}

LOOP:
	for {
		output, err := a.Client.ListShardsWithContext(ctx, params)
		if err != nil {
			return 0, fmt.Errorf("listshards stream %s: %v", stream, err)
		}

		for _, s := range output.Shards {
			if end := s.SequenceNumberRange.EndingSequenceNumber; end == nil {
				shards++
			}
		}

		if output.NextToken != nil {
			params = &kinesis.ListShardsInput{
				NextToken: output.NextToken,
			}
		} else {
			break LOOP
		}
	}

	return shards, nil
}

// UpdateShards uniformly updates a Kinesis stream's shard count and returns when the update is complete.
func (a *API) UpdateShards(ctx aws.Context, stream string, shards int64) error {
	params := &kinesis.UpdateShardCountInput{
		StreamName:       aws.String(stream),
		TargetShardCount: aws.Int64(shards),
		ScalingType:      aws.String("UNIFORM_SCALING"),
	}
	if _, err := a.Client.UpdateShardCountWithContext(ctx, params); err != nil {
		return fmt.Errorf("updateshards stream %s shards %d: %v", stream, shards, err)
	}

	for {
		resp, err := a.Client.DescribeStreamSummaryWithContext(ctx,
			&kinesis.DescribeStreamSummaryInput{
				StreamName: aws.String(stream),
			})
		if err != nil {
			return fmt.Errorf("describestream stream %s: %v", stream, err)
		}

		if status := resp.StreamDescriptionSummary.StreamStatus; status != aws.String("UPDATING") {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// GetTags recursively retrieves all tags for a Kinesis stream.
func (a *API) GetTags(ctx aws.Context, stream string) ([]*kinesis.Tag, error) {
	var tags []*kinesis.Tag
	var lastTag string

	for {
		req := &kinesis.ListTagsForStreamInput{
			StreamName: aws.String(stream),
		}

		if lastTag != "" {
			req.ExclusiveStartTagKey = aws.String(lastTag)
		}

		resp, err := a.Client.ListTagsForStreamWithContext(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("listtags stream %s: %v", stream, err)
		}

		tags = append(tags, resp.Tags...)
		lastTag = *resp.Tags[len(resp.Tags)-1].Key

		// enables recursion
		if !*resp.HasMoreTags {
			break
		}
	}

	return tags, nil
}

// UpdateTag updates a tag on a Kinesis stream.
func (a *API) UpdateTag(ctx aws.Context, stream, key, value string) error {
	input := &kinesis.AddTagsToStreamInput{
		StreamName: aws.String(stream),
		Tags: map[string]*string{
			key: aws.String(value),
		},
	}

	if _, err := a.Client.AddTagsToStreamWithContext(ctx, input); err != nil {
		return fmt.Errorf("updatetag stream %s key %s value %s: %v", stream, key, value, err)
	}

	return nil
}
