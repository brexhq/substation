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
errSQSMessageSizeLimit is returned when data exceeds the SQS message
size limit. If this error occurs, then conditions or processors
should be applied to either drop or reduce the size of the data.
*/
var errSQSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

// awsSQS sinks data to an AWS SQS queue.
type sinkAWSSQS struct {
	// Queue is the AWS SQS queue name that data is sent to.
	Queue string `json:"queue"`
}

// Create a new AWS SQS sink.
func newSinkAWSSQS(_ context.Context, cfg config.Config) (s sinkAWSSQS, err error) {
	if err = config.Decode(cfg.Settings, &s); err != nil {
		return sinkAWSSQS{}, err
	}

	if s.Queue == "" {
		return sinkAWSSQS{}, fmt.Errorf("sink: aws_sqs: queue: %v", errors.ErrMissingRequiredOption)
	}

	return s, nil
}

// Send sinks a channel of encapsulated data with the sink.
func (s sinkAWSSQS) Send(ctx context.Context, ch *config.Channel) error {
	if !sqsAPI.IsEnabled() {
		sqsAPI.Setup()
	}

	// SQS limits messages (both individual and batched)
	// at 256 KB. this buffer will not exceed 256 KB or
	// 500 messages.
	buffer := aggregate.Bytes{}
	buffer.New(500, sqsMessageSizeLimit)

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(capsule.Data()) > sqsMessageSizeLimit {
				return fmt.Errorf("sink: aws_sqs: %v", errSQSMessageSizeLimit)
			}

			ok := buffer.Add(capsule.Data())
			if !ok {
				items := buffer.Get()
				_, err := sqsAPI.SendMessageBatch(ctx, items, s.Queue)
				if err != nil {
					return fmt.Errorf("sink: aws_sqs: %v", err)
				}

				log.WithField(
					"queue", s.Queue,
				).WithField(
					"count", buffer.Count(),
				).Debug("sent messages to SQS")

				buffer.Reset()
				_ = buffer.Add(capsule.Data())
			}
		}
	}

	// send remaining items in buffer
	if buffer.Count() > 0 {
		items := buffer.Get()
		_, err := sqsAPI.SendMessageBatch(ctx, items, s.Queue)
		if err != nil {
			return fmt.Errorf("sink: aws_sqs: %v", err)
		}

		log.WithField(
			"queue", s.Queue,
		).WithField(
			"count", buffer.Count(),
		).Debug("sent messages to SQS")
	}

	return nil
}
