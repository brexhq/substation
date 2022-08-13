package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/kinesis_firehose"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
	"github.com/jshlbrd/go-aggregate"
)

// records greater than 1000 KB in size cannot be
// put into Kinesis Firehose
const kinesisFirehoseRecordSizeLimit = 1000 * 1000

/*
KinesisFirehoseDataExceededSizeLimit is returned when data
exceeds the Kinesis Firehose record size limit. If this
error occurs, then conditions or processors should be
applied to either drop or reduce the size of the data.
*/
const kinesisFirehoseDataExceededSizeLimit = errors.Error("kinesisFirehoseDataExceededSizeLimit")

/*
KinesisFirehose sinks data to an AWS Kinesis Firehose Delivery Stream.
This sink uploads data in batches of records and will automatically retry
any failed put record attempts.

The sink has these settings:
	Stream:
		Kinesis Firehose Delivery Stream that data is sent to

The sink uses this Jsonnet configuration:
	{
		type: 'kinesis_firehose',
		settings: {
			stream: 'foo',
		},
	}
*/
type KinesisFirehose struct {
	Stream string `json:"stream"`
}

var kinesisFirehoseAPI kinesis_firehose.API

// Send sinks a channel of encapsulated data with the Kinesis sink.
func (sink *KinesisFirehose) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	if !kinesisFirehoseAPI.IsEnabled() {
		kinesisFirehoseAPI.Setup()
	}

	// Kinesis Firehose limits Batch operations at up to 4 MB
	// and 500 records per batch. this buffer will not exceed
	// 3.6 MB or 500 records.
	buffer := aggregate.Bytes{}
	buffer.New(kinesisFirehoseRecordSizeLimit*.9*4, 500)

	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			if len(cap.GetData()) > kinesisFirehoseRecordSizeLimit {
				return fmt.Errorf("sink kinesis firehose: %v", kinesisFirehoseDataExceededSizeLimit)
			}

			ok, err := buffer.Add(cap.GetData())
			if err != nil {
				return fmt.Errorf("sink kinesis firehose: %v", err)
			}

			if !ok {
				items := buffer.Get()
				_, err := kinesisFirehoseAPI.PutRecordBatch(ctx, items, sink.Stream)
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
		_, err := kinesisFirehoseAPI.PutRecordBatch(ctx, items, sink.Stream)
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
