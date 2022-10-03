package firehose

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// New creates a new session for Kinesis Firehose
func New() *firehose.Firehose {
	conf := aws.NewConfig()

	// provides forward compatibility for the Go SDK to support env var configuration settings
	// https://github.com/aws/aws-sdk-go/issues/4207
	max, found := os.LookupEnv("AWS_MAX_ATTEMPTS")
	if found {
		m, err := strconv.Atoi(max)
		if err != nil {
			panic(err)
		}

		conf = conf.WithMaxRetries(m)
	}

	c := firehose.New(
		session.Must(session.NewSession()),
		conf,
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps a Kinesis Firehose client interface
type API struct {
	Client firehoseiface.FirehoseAPI
}

// IsEnabled checks whether a new client has been set
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// Setup creates a Kinesis Firehose client
func (a *API) Setup() {
	a.Client = New()
}

// PutRecord is a convenience wrapper for putting a record into a Kinesis Firehose stream.
func (a *API) PutRecord(ctx aws.Context, data []byte, stream string) (*firehose.PutRecordOutput, error) {
	resp, err := a.Client.PutRecordWithContext(
		ctx,
		&firehose.PutRecordInput{
			DeliveryStreamName: aws.String(stream),
			Record:             &firehose.Record{Data: data},
		})
	if err != nil {
		return nil, fmt.Errorf("putrecord stream %s: %v", stream, err)
	}

	return resp, nil
}

// PutRecordBatch is a convenience wrapper for putting multiple records into a Kinesis Firehose stream. This function becomes recursive for any records that failed the PutRecord operation.
func (a *API) PutRecordBatch(ctx aws.Context, data [][]byte, stream string) (*firehose.PutRecordBatchOutput, error) {
	var records []*firehose.Record
	for _, d := range data {
		records = append(records, &firehose.Record{Data: d})
	}

	resp, err := a.Client.PutRecordBatchWithContext(
		ctx,
		&firehose.PutRecordBatchInput{
			DeliveryStreamName: aws.String(stream),
			Records:            records,
		},
	)

	// failed records are identified by the existence of an error code.
	// if an error code exists, then data is stored in a new slice and
	// recursively input into the function.
	if resp.FailedPutCount != aws.Int64(0) {
		var retryData [][]byte
		for idx, r := range resp.RequestResponses {
			if r.ErrorCode == nil {
				continue
			}

			retryData = append(retryData, data[idx])
		}

		if len(retryData) > 0 {
			_, _ = a.PutRecordBatch(ctx, retryData, stream)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("putrecordbatch stream %s: %v", stream, err)
	}

	return resp, nil
}
