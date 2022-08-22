package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/sqs"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
	"github.com/jshlbrd/go-aggregate"
)

var sqsAPI sqs.API

// records greater than 256 KB in size cannot be
// put into an SQS queue
const sqsMessageSizeLimit = 1024 * 1024 * 256

/*
SQSDataExceededSizeLimit is returned when data exceeds
the SQS message size limit. If this error occurs, then
conditions or processors should be applied to either drop
or reduce the size of the data.
*/
const SQSDataExceededSizeLimit = errors.Error("KinesisFirehoseDataExceededSizeLimit")

/*
SQS sinks data to an AWS SQS queue.

The sink has these settings:
	Queue:
		SQS queue name that data is sent to

When loaded with a factory, the sink uses this JSON configuration:
	{
		"type": "sqs",
		"settings": {
			"queue": "foo"
		}
	}
*/
type SQS struct {
	Queue string `json:"queue"`
}

// Send sinks a channel of encapsulated data with the Kinesis sink.
func (sink *SQS) Send(ctx context.Context, ch chan config.Capsule, kill chan struct{}) error {
	if !sqsAPI.IsEnabled() {
		sqsAPI.Setup()
	}

	// SQS limits messages (both individual and batched)
	// at 256 KB. this buffer will not exceed 256 KB or
	// 500 messages.
	buffer := aggregate.Bytes{}
	buffer.New(sqsMessageSizeLimit, 500)

	for cap := range ch {
		select {
		case <-kill:
			return nil
		default:
			if len(cap.GetData()) > sqsMessageSizeLimit {
				return fmt.Errorf("sink sqs: %v", SQSDataExceededSizeLimit)
			}

			ok, err := buffer.Add(cap.GetData())
			if err != nil {
				return fmt.Errorf("sink sqs: %v", err)
			}

			if !ok {
				items := buffer.Get()
				_, err := sqsAPI.SendMessageBatch(ctx, items, sink.Queue)
				if err != nil {
					return fmt.Errorf("sink sqs: %v", err)
				}

				log.WithField(
					"queue", sink.Queue,
				).WithField(
					"count", buffer.Count(),
				).Debug("sent messages to SQS")

				buffer.Reset()
				buffer.Add(cap.GetData())
			}
		}
	}

	// send remaining items in buffer
	if buffer.Count() > 0 {
		items := buffer.Get()
		_, err := sqsAPI.SendMessageBatch(ctx, items, sink.Queue)
		if err != nil {
			return fmt.Errorf("sink sqs: %v", err)
		}

		log.WithField(
			"queue", sink.Queue,
		).WithField(
			"count", buffer.Count(),
		).Debug("sent messages to SQS")
	}

	return nil
}
