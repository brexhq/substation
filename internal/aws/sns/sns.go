package sns

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/google/uuid"
)

// New creates a new session for SNS.
func New() *sns.SNS {
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

	c := sns.New(
		session.Must(session.NewSession()),
		conf,
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps an SNS client interface.
type API struct {
	Client snsiface.SNSAPI
}

// IsEnabled checks whether a new client has been set.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// Setup creates an SNS client.
func (a *API) Setup() {
	a.Client = New()
}

// Publish is a convenience wrapper for publishing a message to an SNS topic.
func (a *API) Publish(ctx aws.Context, data []byte, topic string) (*sns.PublishOutput, error) {
	mgid := uuid.New().String()

	resp, err := a.Client.PublishWithContext(
		ctx,
		&sns.PublishInput{
			Message:        aws.String(string(data)),
			MessageGroupId: aws.String(mgid),
			TopicArn:       aws.String(topic),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("publish topic %s: %v", topic, err)
	}

	return resp, nil
}

// PublishBatch is a convenience wrapper for publishing a batch of messages to an SNS topic.
func (a *API) PublishBatch(ctx aws.Context, data [][]byte, topic string, fifo bool) (*sns.PublishBatchOutput, error) {
	mgid := uuid.New().String()

	var entries []*sns.PublishBatchRequestEntry
	for idx, d := range data {
		entry := &sns.PublishBatchRequestEntry{
			Id:      aws.String(strconv.Itoa(idx)),
			Message: aws.String(string(d)),
		}

		if fifo {
			entry.MessageGroupId = aws.String(mgid)
		}

		entries = append(entries, entry)
	}

	resp, err := a.Client.PublishBatchWithContext(
		ctx,
		&sns.PublishBatchInput{
			PublishBatchRequestEntries: entries,
			TopicArn:                   aws.String(topic),
		},
	)

	// if a message fails, then the message ID is used to select the
	// original data that was in the message. this data is put in a
	// new slice and recursively input into the function.
	if resp.Failed != nil {
		var retry [][]byte
		for _, f := range resp.Failed {
			idx, err := strconv.Atoi(*f.Id)
			if err != nil {
				return nil, err
			}

			retry = append(retry, data[idx])
		}

		if len(retry) > 0 {
			_, _ = a.PublishBatch(ctx, retry, topic, fifo)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("publish batch topic %s: %v", topic, err)
	}

	return resp, nil
}
