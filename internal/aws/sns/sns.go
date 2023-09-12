package sns

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	iaws "github.com/brexhq/substation/internal/aws"
	"github.com/google/uuid"
)

// New returns a configured SNS client.
func New(cfg iaws.Config) *sns.SNS {
	conf, sess := iaws.New(cfg)

	c := sns.New(sess, conf)
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
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
}

// Publish is a convenience wrapper for publishing a message to an SNS topic.
func (a *API) Publish(ctx aws.Context, arn string, data []byte) (*sns.PublishOutput, error) {
	mgid := uuid.New().String()

	resp, err := a.Client.PublishWithContext(
		ctx,
		&sns.PublishInput{
			Message:        aws.String(string(data)),
			MessageGroupId: aws.String(mgid),
			TopicArn:       aws.String(arn),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("publish: topic %s: %v", arn, err)
	}

	return resp, nil
}

// PublishBatch is a convenience wrapper for publishing a batch of messages to an SNS topic.
func (a *API) PublishBatch(ctx aws.Context, topic string, data [][]byte) (*sns.PublishBatchOutput, error) {
	mgid := uuid.New().String()

	var entries []*sns.PublishBatchRequestEntry
	for idx, d := range data {
		entry := &sns.PublishBatchRequestEntry{
			Id:      aws.String(strconv.Itoa(idx)),
			Message: aws.String(string(d)),
		}

		if strings.HasSuffix(topic, ".fifo") {
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
			return a.PublishBatch(ctx, topic, retry)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("publish_batch: topic %s: %v", topic, err)
	}

	return resp, nil
}
