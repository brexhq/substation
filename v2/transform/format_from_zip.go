package transform

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/message"
)

type formatZipConfig struct {
	ID string `json:"id"`
}

func (c *formatZipConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newFormatFromZip(_ context.Context, cfg config.Config) (*formatFromZip, error) {
	conf := formatZipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_from_zip: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_from_zip"
	}

	tf := formatFromZip{
		conf: conf,
	}

	return &tf, nil
}

type formatFromZip struct {
	conf formatZipConfig
}

func (tf *formatFromZip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	b := bytes.NewReader(msg.Data())
	r, err := zip.NewReader(b, int64(len(msg.Data())))
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	var msgs []*message.Message
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if f.FileInfo().Size() == 0 {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
		defer rc.Close()

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(rc); err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		m := message.New().SetData(buf.Bytes()).SetMetadata(msg.Metadata())
		msgs = append(msgs, m)
	}

	return msgs, nil
}

func (tf *formatFromZip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
