package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/lambda"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

// Payloads greater than 256 KB in size cannot be
// sent to an AWS Lambda function.
const sendLambdaPayloadSizeLimit = 1024 * 1024 * 256

// errSendLambdaPayloadSizeLimit is returned when data exceeds the Lambda
// payload size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendLambdaPayloadSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSLambdaConfig struct {
	// FunctionName is the AWS Lambda function to asynchronously invoke.
	FunctionName string `json:"function_name"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
	Retry  iconfig.Retry  `json:"retry"`
}

func (c *sendAWSLambdaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSLambdaConfig) Validate() error {
	if c.FunctionName == "" {
		return fmt.Errorf("function_name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSLambda(_ context.Context, cfg config.Config) (*sendAWSLambda, error) {
	conf := sendAWSLambdaConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_lambda: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_lambda"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSLambda{
		conf:     conf,
		function: conf.FunctionName,
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    conf.Batch.Count,
		Size:     conf.Batch.Size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.agg = agg

	if len(conf.AuxTransforms) > 0 {
		tf.tforms = make([]Transformer, len(conf.AuxTransforms))
		for i, c := range conf.AuxTransforms {
			t, err := New(context.Background(), c)
			if err != nil {
				return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
			}

			tf.tforms[i] = t
		}
	}

	// Setup the AWS client.
	tf.client.Setup(aws.Config{
		Region:          conf.AWS.Region,
		RoleARN:         conf.AWS.RoleARN,
		MaxRetries:      conf.Retry.Count,
		RetryableErrors: conf.Retry.ErrorMessages,
	})

	return &tf, nil
}

type sendAWSLambda struct {
	conf     sendAWSLambdaConfig
	function string

	// client is safe for concurrent use.
	client lambda.API

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSLambda) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		for key := range tf.agg.GetAll() {
			if tf.agg.Count(key) == 0 {
				continue
			}

			if err := tf.send(ctx, key); err != nil {
				return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
			}
		}

		tf.agg.ResetAll()
		return []*message.Message{msg}, nil
	}

	if len(msg.Data()) > sendLambdaPayloadSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendLambdaPayloadSizeLimit)
	}

	// If this value does not exist, then all data is batched together.
	key := msg.GetValue(tf.conf.Object.BatchKey).String()
	if ok := tf.agg.Add(key, msg.Data()); ok {
		return []*message.Message{msg}, nil
	}

	if err := tf.send(ctx, key); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If data cannot be added after reset, then the batch is misconfgured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendBatchMisconfigured)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSLambda) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	for _, b := range data {
		if _, err := tf.client.InvokeAsync(ctx, tf.function, b); err != nil {
			return err
		}
	}

	return nil
}

func (tf *sendAWSLambda) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
