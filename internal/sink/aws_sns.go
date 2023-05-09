package sink

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/sns"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/log"
	"github.com/jshlbrd/go-aggregate"
)

var snsAPI sns.API

// records greater than 256 KB in size cannot be
// put into an SNS topic
const snsMessageSizeLimit = 1024 * 1024 * 256

/*
errSNSMessageSizeLimit is returned when data exceeds the SNS message
size limit. If this error occurs, then conditions or processors
should be applied to either drop or reduce the size of the data.
*/
var errSNSMessageSizeLimit = fmt.Errorf("data exceeded size limit")

// awsSNS sinks data to an AWS SNS topic.
type sinkAWSSNS struct {
	// ARN is the ARN of the AWS SNS topic that data is sent to.
	ARN string `json:"arn"`
}

// Create a new AWS SNS sink.
func newSinkAWSSNS(_ context.Context, cfg config.Config) (s sinkAWSSNS, err error) {
	if err = config.Decode(cfg.Settings, &s); err != nil {
		return sinkAWSSNS{}, err
	}

	if s.ARN == "" {
		return sinkAWSSNS{}, fmt.Errorf("sink: aws_sns: arn: %v", errors.ErrMissingRequiredOption)
	}

	return s, nil
}

// Send sinks a channel of encapsulated data with the sink.
func (s sinkAWSSNS) Send(ctx context.Context, ch *config.Channel) error {
	if !snsAPI.IsEnabled() {
		snsAPI.Setup()
	}

	// SNS limits messages (both individual and batched)
	// at 256 KB. this buffer will not exceed 256 KB or
	// 10 messages.
	buffer := aggregate.Bytes{}
	buffer.New(10, snsMessageSizeLimit)

	for capsule := range ch.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(capsule.Data()) > snsMessageSizeLimit {
				return fmt.Errorf("sink: aws_sns: %v", errSNSMessageSizeLimit)
			}

			ok := buffer.Add(capsule.Data())
			if !ok {
				items := buffer.Get()
				_, err := snsAPI.PublishBatch(ctx, s.ARN, items)
				if err != nil {
					return fmt.Errorf("sink: aws_sns: %v", err)
				}

				log.WithField(
					"arn", s.ARN,
				).WithField(
					"count", buffer.Count(),
				).Debug("sent messages to SNS")

				buffer.Reset()
				_ = buffer.Add(capsule.Data())
			}
		}
	}

	// send remaining items in buffer
	if buffer.Count() > 0 {
		items := buffer.Get()
		_, err := snsAPI.PublishBatch(ctx, s.ARN, items)
		if err != nil {
			return fmt.Errorf("sink: aws_sns: %v", err)
		}

		log.WithField(
			"arn", s.ARN,
		).WithField(
			"count", buffer.Count(),
		).Debug("sent messages to SNS")
	}

	return nil
}
