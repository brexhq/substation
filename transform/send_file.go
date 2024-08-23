package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/internal/aggregate"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/file"
	"github.com/brexhq/substation/v2/message"
)

type sendFileConfig struct {
	// FilePath determines how the name of the file is constructed.
	// See filePath.New for more information.
	FilePath file.Path `json:"file_path"`
	// UseBatchKeyAsPrefix determines if the batch key should be used as the prefix.
	UseBatchKeyAsPrefix bool `json:"use_batch_key_as_prefix"`
	// AuxTransforms are applied to batched data before it is sent.
	AuxTransforms []config.Config `json:"auxiliary_transforms"`

	ID     string         `json:"id"`
	Object iconfig.Object `json:"object"`
	Batch  iconfig.Batch  `json:"batch"`
}

func (c *sendFileConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *sendFileConfig) Validate() error {
	return nil
}

func newSendFile(_ context.Context, cfg config.Config) (*sendFile, error) {
	conf := sendFileConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_file: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_file"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := sendFile{
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

	return &tf, nil
}

type sendFile struct {
	conf sendFileConfig

	mu     sync.Mutex
	agg    *aggregate.Aggregate
	tforms []Transformer
}

func (tf *sendFile) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

func (tf *sendFile) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (tf *sendFile) send(ctx context.Context, key string) error {
	p := tf.conf.FilePath
	if key != "" && tf.conf.UseBatchKeyAsPrefix {
		p.Prefix = key
	}

	path := p.New()
	if path == "" {
		return fmt.Errorf("file path is empty")
	}

	// Ensures that the path is OS agnostic.
	path = filepath.FromSlash(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o770); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	data, err := withTransforms(ctx, tf.tforms, tf.agg.Get(key))
	if err != nil {
		return err
	}

	for _, d := range data {
		if _, err := f.Write(d); err != nil {
			return err
		}
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
