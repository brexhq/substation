package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

// Records greater than 1000 KiB in size cannot be put into Kinesis Firehose.
const sendAWSDataFirehoseMessageSizeLimit = 1024 * 1000

// errSendAWSDataFirehoseRecordSizeLimit is returned when data exceeds the
// Kinesis Firehose record size limit. If this error occurs,
// then drop or reduce the size of the data before attempting to
// send it to Kinesis Firehose.
var errSendAWSDataFirehoseRecordSizeLimit = fmt.Errorf("data exceeded size limit")

type sendAWSDataFirehoseConfig struct {
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *sendAWSDataFirehoseConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSDataFirehoseConfig) Validate() error {
	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newSendAWSDataFirehose(ctx context.Context, cfg config.Config) (*sendAWSDataFirehose, error) {
	conf := sendAWSDataFirehoseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_kinesis_data_firehose: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_kinesis_data_firehose"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSDataFirehose{
		conf: conf,
	}

	awsCfg, err := iconfig.NewAWS(ctx, conf.AWS)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.client = firehose.NewFromConfig(awsCfg)

	// Data Firehose limits batch operations to 500 records.
	count := 500
	if conf.Batch.Count > 0 && conf.Batch.Count <= count {
		count = conf.Batch.Count
	}

	// Data Firehose limits batch operations to 4 MiB.
	size := sendAWSDataFirehoseMessageSizeLimit * 4
	if conf.Batch.Size > 0 && conf.Batch.Size <= size {
		size = conf.Batch.Size
	}

	agg, err := aggregate.New(aggregate.Config{
		Count:    count,
		Size:     size,
		Duration: conf.Batch.Duration,
	})
	if err != nil {
		return nil, err
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

	return &tf, nil
}

type sendAWSDataFirehose struct {
	conf   sendAWSDataFirehoseConfig
	client *firehose.Client

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSDataFirehose) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	if len(msg.Data()) > sendAWSDataFirehoseMessageSizeLimit {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errSendAWSDataFirehoseRecordSizeLimit)
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
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendAWSDataFirehose) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSDataFirehose) send(ctx context.Context, key string) error {
	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	ctx = context.WithoutCancel(ctx)
	return tf.putRecords(ctx, data)
}

func (tf *sendAWSDataFirehose) putRecords(ctx context.Context, data [][]byte) error {
	var records []types.Record
	for _, d := range data {
		records = append(records, types.Record{
			Data: d,
		})
	}

	resp, err := tf.client.PutRecordBatch(ctx, &firehose.PutRecordBatchInput{
		DeliveryStreamName: &tf.conf.AWS.ARN,
		Records:            records,
	})
	if resp.FailedPutCount != nil && *resp.FailedPutCount > 0 {
		var retry [][]byte

		for i, r := range resp.RequestResponses {
			if r.ErrorCode != nil {
				retry = append(retry, data[i])
			}
		}

		if len(retry) > 0 {
			return tf.putRecords(ctx, retry)
		}
	}

	if err != nil {
		return err
	}

	return nil
}
