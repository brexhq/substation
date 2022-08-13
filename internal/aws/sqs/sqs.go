package sqs

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// New creates a new session for SQS
func New() *sqs.SQS {
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

	c := sqs.New(
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
	Client sqsiface.SQSAPI
}

// IsEnabled checks whether a new client has been set
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// Setup creates a Kinesis Firehose client
func (a *API) Setup() {
	a.Client = New()
}

// SendMessage is a convenience wrapper for sending a message to an SQS queue.
func (a *API) SendMessage(ctx aws.Context, data []byte, queue string) (*sqs.SendMessageOutput, error) {
	resp, err := a.Client.SendMessageWithContext(
		ctx,
		&sqs.SendMessageInput{
			MessageBody: aws.String(string(data)),
			QueueUrl:    aws.String(queue),
		})

	if err != nil {
		return nil, fmt.Errorf("sendmessage queue %s: %w", queue, err)
	}

	return resp, nil
}

// SendMessageBatch is a convenience wrapper for sending multiple messages to an SQS queue. This function becomes recursive for any messages that failed the SendMessage operation.
func (a *API) SendMessageBatch(ctx aws.Context, data [][]byte, queue string) (*sqs.SendMessageBatchOutput, error) {
	var messages []*sqs.SendMessageBatchRequestEntry
	for idx, d := range data {
		messages = append(messages, &sqs.SendMessageBatchRequestEntry{
			Id:          aws.String(strconv.Itoa(idx)),
			MessageBody: aws.String(string(d)),
		})
	}

	url, err := a.Client.GetQueueUrlWithContext(
		ctx,
		&sqs.GetQueueUrlInput{
			QueueName: aws.String(queue),
		})
	if err != nil {
		return nil, fmt.Errorf("sendmessagebatch queue %s: %w", queue, err)
	}

	resp, err := a.Client.SendMessageBatchWithContext(
		ctx,
		&sqs.SendMessageBatchInput{
			Entries:  messages,
			QueueUrl: url.QueueUrl,
		},
	)

	// if a message fails, then the message ID is used to select the
	// original data that was in the message. this data is put in a
	// new slice and recursively input into the function.
	if resp.Failed != nil {
		var retryData [][]byte
		for _, r := range resp.Failed {
			idx, err := strconv.Atoi(aws.StringValue(r.Id))
			if err != nil {
				return nil, fmt.Errorf("sendmessagebatch queue %s: %w", queue, err)
			}

			retryData = append(retryData, data[idx])
		}

		if len(retryData) > 0 {
			a.SendMessageBatch(ctx, retryData, queue)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("sendmessagebatch queue %s: %w", queue, err)
	}

	return resp, nil
}
