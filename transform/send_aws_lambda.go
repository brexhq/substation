package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iaggregate "github.com/brexhq/substation/v2/internal/aggregate"
	iaws "github.com/brexhq/substation/v2/internal/aws"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	ierrors "github.com/brexhq/substation/v2/internal/errors"
)

// Payloads greater than 256 KB in size cannot be
// sent to an AWS Lambda function.
const sendLambdaPayloadSizeLimit = 1024 * 1024 * 256

// errSendLambdaPayloadSizeLimit is returned when data exceeds the Lambda
// payload size limit. If this error occurs, then conditions or transforms
// should be applied to either drop or reduce the size of the data.
var errSendLambdaPayloadSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSLambdaConfig struct {
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *sendAWSLambdaConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSLambdaConfig) Validate() error {
	if c.AWS.ARN == "" {
		return fmt.Errorf("arn: %v", ierrors.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSLambda(ctx context.Context, cfg config.Config) (*sendAWSLambda, error) {
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
		conf: conf,
	}

	agg, err := iaggregate.New(iaggregate.Config{
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
	awsCfg, err := iaws.New(ctx, iaws.Config{
		Region:  iaws.ParseRegion(conf.AWS.ARN),
		RoleARN: conf.AWS.AssumeRoleARN,
	})
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = lambda.NewFromConfig(awsCfg)

	return &tf, nil
}

type sendAWSLambda struct {
	conf   sendAWSLambdaConfig
	client *lambda.Client

	mu     sync.Mutex
	agg    *iaggregate.Aggregate
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

	input := &lambda.InvokeInput{
		FunctionName:   &tf.conf.AWS.ARN,
		InvocationType: "Event", // Asynchronous invocation.
	}

	ctx = context.WithoutCancel(ctx)
	for _, d := range data {
		input.Payload = d
		if _, err := tf.client.Invoke(ctx, input); err != nil {
			return err
		}
	}

	return nil
}

func (tf *sendAWSLambda) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
