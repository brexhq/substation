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
FirehoseDataExceededSizeLimit is returned when data
exceeds the Kinesis Firehose record size limit. If this
error occurs, then conditions or processors should be
applied to either drop or reduce the size of the data.
*/
const FirehoseDataExceededSizeLimit = errors.Error("FirehoseDataExceededSizeLimit")

/*
Firehose sinks data to an AWS Kinesis Firehose Delivery Stream.
This sink uploads data in batches of records and will automatically retry
any failed put record attempts.

The sink has these settings:
	Stream:
		Kinesis Firehose Delivery Stream that data is sent to

When loaded with a factory, the sink uses this JSON configuration:
	{
		"type": "firehose",
		"settings": {
			"stream": "foo"
		}
	}
*/
type Firehose struct {
	Stream string `json:"stream"`
}

// Send sinks a channel of encapsulated data with the Kinesis sink.
func (sink *Firehose) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	if !firehoseAPI.IsEnabled() {
		firehoseAPI.Setup()
	}

	// Firehose limits Batch operations at up to 4 MiB and
	// 500 records per batch. this buffer will not exceed
	// 3.9 MiB or 500 records.
	buffer := aggregate.Bytes{}
	buffer.New(firehoseRecordSizeLimit*4*.99, 500)

	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			if len(cap.GetData()) > firehoseRecordSizeLimit {
				return fmt.Errorf("sink kinesis firehose: %v", FirehoseDataExceededSizeLimit)
			}

			ok, err := buffer.Add(cap.GetData())
			if err != nil {
				return fmt.Errorf("sink kinesis firehose: %v", err)
			}

			if !ok {
				items := buffer.Get()
				_, err := firehoseAPI.PutRecordBatch(ctx, items, sink.Stream)
				if err != nil {
					return fmt.Errorf("sink kinesis firehose: %v", err)
				}

				log.WithField(
					"stream", sink.Stream,
				).WithField(
					"count", buffer.Count(),
				).Debug("put records into Kinesis Firehose")

				buffer.Reset()
				buffer.Add(cap.GetData())
			}
		}
	}

	// send remaining items in buffer
	if buffer.Count() > 0 {
		items := buffer.Get()
		_, err := firehoseAPI.PutRecordBatch(ctx, items, sink.Stream)
		if err != nil {
			return fmt.Errorf("sink kinesis firehose: %v", err)
		}

		log.WithField(
			"stream", sink.Stream,
		).WithField(
			"count", buffer.Count(),
		).Debug("put records into Kinesis Firehose")
	}

	return nil
}
