package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/file"
	"github.com/brexhq/substation/v2/internal/media"
)

type sendAWSS3Config struct {
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
	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	if types.StorageClass(c.StorageClass) == "" {
		return fmt.Errorf("storage class: %v", iconfig.ErrInvalidOption)
	}

	return nil
}

func newSendAWSS3(ctx context.Context, cfg config.Config) (*sendAWSS3, error) {
	conf := sendAWSS3Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_aws_s3: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_aws_s3"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendAWSS3{
		conf: conf,
	}

	// Extracts the bucket name from the ARN.
	// The ARN is in the format: arn:aws:s3:::bucket-name
	a, err := arn.Parse(conf.AWS.ARN)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf.bucket = a.Resource

	if conf.StorageClass == "" {
		tf.sclass = types.StorageClassStandard
	} else {
		tf.sclass = types.StorageClass(conf.StorageClass)
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

	awsCfg, err := iconfig.NewAWS(ctx, conf.AWS)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	c := s3.NewFromConfig(awsCfg)
	tf.client = manager.NewUploader(c)

	return &tf, nil
}

type sendAWSS3 struct {
	conf   sendAWSS3Config
	client *manager.Uploader
	bucket string
	sclass types.StorageClass

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
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
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

	mediaType, err := media.File(f)
	if err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	ctx = context.WithoutCancel(ctx)
	if _, err := tf.client.Upload(ctx, &s3.PutObjectInput{
		Bucket:       &tf.bucket,
		Key:          &filePath,
		Body:         f,
		StorageClass: tf.sclass,
		ContentType:  &mediaType,
	}); err != nil {
		return err
	}

	return nil
}
