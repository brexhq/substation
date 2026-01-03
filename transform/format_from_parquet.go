package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/parquet-go/parquet-go"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type formatParquetConfig struct {
	ID string `json:"id"`
}

func (c *formatParquetConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newFormatFromParquet(_ context.Context, cfg config.Config) (*formatFromParquet, error) {
	conf := formatParquetConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_from_parquet: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_from_parquet"
	}

	tf := formatFromParquet{
		conf: conf,
	}

	return &tf, nil
}

type formatFromParquet struct {
	conf formatParquetConfig
}

func (tf *formatFromParquet) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	fi, err := parquet.OpenFile(bytes.NewReader(msg.Data()), int64(len(msg.Data())))
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	reader := parquet.NewGenericReader[any](fi)
	defer reader.Close()

	rows := make([]any, reader.NumRows())
	for {
		if n, err := reader.Read(rows); err != nil {
			if err.Error() == "EOF" || n == 0 {
				break
			}

			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
	}

	var msgs []*message.Message
	for _, row := range rows {
		data, err := json.Marshal(row)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		msgs = append(msgs, message.New().SetData(data).SetMetadata(msg.Metadata()))
	}

	return msgs, nil
}

func (tf *formatFromParquet) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
