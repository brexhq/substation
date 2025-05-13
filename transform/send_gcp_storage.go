package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/storage"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/file"
	"github.com/brexhq/substation/v2/internal/media"
)

type sendGCPStorageConfig struct {
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
	GCP    iconfig.GCP    `json:"gcp"`
}

func (c *sendGCPStorageConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendGCPStorageConfig) Validate() error {
	if c.GCP.Resource == "" {
		return fmt.Errorf("gcp.resource: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newSendGCPStorage(ctx context.Context, cfg config.Config) (*sendGCPStorage, error) {
	conf := sendGCPStorageConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_gcp_storage: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_gcp_storage"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendGCPStorage{
		conf:   conf,
		bucket: conf.GCP.Resource,
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

	// Initialize GCP Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf.client = client

	return &tf, nil
}

type sendGCPStorage struct {
	conf   sendGCPStorageConfig
	client *storage.Client
	bucket string

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendGCPStorage) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	// If data cannot be added after reset, then the batch is misconfigured.
	tf.agg.Reset(key)
	if ok := tf.agg.Add(key, msg.Data()); !ok {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errBatchNoMoreData)
	}

	return []*message.Message{msg}, nil
}

func (tf *sendGCPStorage) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendGCPStorage) send(ctx context.Context, key string) error {
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

	// Flush the file before uploading to GCS.
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
	obj := tf.client.Bucket(tf.bucket).Object(filePath)
	writer := obj.NewWriter(ctx)

	// Set object attributes
	writer.ContentType = mediaType

	if _, err := writer.Write(data[0]); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}
