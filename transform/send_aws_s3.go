package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	"github.com/brexhq/substation/v2/internal/aws"
	"github.com/brexhq/substation/v2/internal/aws/s3manager"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/internal/file"
)

type sendAWSS3Config struct {
	// BucketName is the AWS S3 bucket that data is written to.
	BucketName string `json:"bucket_name"`
	// StorageClass is the storage class of the object.
	StorageClass string `json:"storage_class"`
	// FilePath determines how the name of the uploaded object is constructed.
	// See filePath.New for more information.
	FilePath file.Path `json:"file_path"`
	// UseBatchKeyAsPrefix determines if the batch key should be used as the prefix.
	UseBatchKeyAsPrefix bool `json:"use_batch_key_as_prefix"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
	AWS    iconfig.AWS    `json:"aws"`
}

func (c *sendAWSS3Config) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendAWSS3Config) Validate() error {
	if c.BucketName == "" {
		return fmt.Errorf("bucket_name: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(s3.StorageClass_Values(), c.StorageClass) {
		return fmt.Errorf("storage_class: %v", errors.ErrInvalidOption)
	}

	return nil
}

func newSendAWSS3(_ context.Context, cfg config.Config) (*sendAWSS3, error) {
	conf := sendAWSS3Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_s3: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_s3"
	}

	if conf.StorageClass == "" {
		conf.StorageClass = "STANDARD"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSS3{
		conf: conf,
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
		Region:  conf.AWS.Region,
		RoleARN: conf.AWS.RoleARN,
	})

	return &tf, nil
}

type sendAWSS3 struct {
	conf sendAWSS3Config

	// client is safe for concurrent use.
	client s3manager.UploaderAPI

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendAWSS3) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

func (tf *sendAWSS3) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendAWSS3) send(ctx context.Context, key string) error {
	p := tf.conf.FilePath
	if key != "" && tf.conf.UseBatchKeyAsPrefix {
		p.Prefix = key
	}

	filePath := p.New()
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	temp, err := os.CreateTemp("", "substation")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	for _, d := range data {
		if _, err := temp.Write(d); err != nil {
			return err
		}
	}

	// Flush the file before uploading to S3.
	if err := temp.Close(); err != nil {
		return err
	}

	f, err := os.Open(temp.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := tf.client.Upload(ctx, tf.conf.BucketName, filePath, tf.conf.StorageClass, f); err != nil {
		return err
	}

	return nil
}
