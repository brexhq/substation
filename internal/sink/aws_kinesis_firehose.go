package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/firehose"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
	"github.com/jshlbrd/go-aggregate"
)

var firehoseAPI firehose.API

// records greater than 1000 KiB in size cannot be
// put into Kinesis Firehose
const firehoseRecordSizeLimit = 1024 * 1000

/*
errFirehoseRecordSizeLimit is returned when data exceeds the
Kinesis Firehose record size limit. If this error occurs,
then conditions or processors should be applied to either
drop or reduce the size of the data.
*/
const errFirehoseRecordSizeLimit = errors.Error("data exceeded size limit")

// awsKinesisFirehose sinks data to an AWS Kinesis Firehose Delivery Stream.
//
// Data is sent in batches of records and will automatically retry
// any failed PutRecord attempts.
type _awsKinesisFirehose struct {
	// Stream is the Kinesis Firehose Delivery Stream that data is sent to.
	Stream string `json:"stream"`
}

// Send sinks a channel of encapsulated data with the sink.
func (sink *_awsKinesisFirehose) Send(ctx context.Context, ch *config.Channel) error {
	if !firehoseAPI.IsEnabled() {
		firehoseAPI.Setup()
	}

	// Firehose limits Batch operations at up to 4 MiB and
	// 500 records per batch. this buffer will not exceed
	// 3.9 MiB or 500 records.
	buffer := aggregate.Bytes{}
	buffer.New(500, firehoseRecordSizeLimit*4*.99)

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(capsule.Data()) > firehoseRecordSizeLimit {
				return fmt.Errorf("sink: aws_kinesis_firehose: %v", errFirehoseRecordSizeLimit)
			}

			ok := buffer.Add(capsule.Data())
			if !ok {
				items := buffer.Get()
				_, err := firehoseAPI.PutRecordBatch(ctx, items, sink.Stream)
				if err != nil {
					return fmt.Errorf("sink: aws_kinesis_firehose: %v", err)
				}

				log.WithField(
					"stream", sink.Stream,
				).WithField(
					"count", buffer.Count(),
				).Debug("put records into Kinesis Firehose")

				buffer.Reset()

				_ = buffer.Add(capsule.Data())
			}
		}
	}

	// send remaining items in buffer
	if buffer.Count() > 0 {
		items := buffer.Get()
		_, err := firehoseAPI.PutRecordBatch(ctx, items, sink.Stream)
		if err != nil {
			return fmt.Errorf("sink: aws_kinesis_firehose: %v", err)
		}

		log.WithField(
			"stream", sink.Stream,
		).WithField(
			"count", buffer.Count(),
		).Debug("put records into Kinesis Firehose")
	}

	return nil
}
