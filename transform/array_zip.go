package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/tidwall/sjson"
)

type arrayZipConfig struct {
	Object   iconfig.Object `json:"object"`
	AsObject bool           `json:"as_object"`
	WithKeys []string       `json:"with_keys"`
}

func (c *arrayZipConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *arrayZipConfig) Validate() error {
	if c.Object.SrcKey == "" {
		return fmt.Errorf("object_src_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.DstKey == "" {
		return fmt.Errorf("object_dst_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newArrayZip(_ context.Context, cfg config.Config) (*arrayZip, error) {
	conf := arrayZipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: array_zip: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: array_zip: %v", err)
	}

	tf := arrayZip{
		conf: conf,
	}

	return &tf, nil
}

type arrayZip struct {
	conf arrayZipConfig
}

func (tf *arrayZip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	value := msg.GetValue(tf.conf.Object.SrcKey)
	if !value.Exists() || !value.IsArray() {
		return []*message.Message{msg}, nil
	}

	switch tf.conf.AsObject {
	case true:
		cache := make(map[int][]message.Value)

		if len(tf.conf.WithKeys) > 0 {
			for i, key := range tf.conf.WithKeys {
				v := bytesToValue([]byte(key))
				cache[i] = append(cache[i], v)
			}
		}

		for _, val := range value.Array() {
			for i, v := range val.Array() {
				cache[i] = append(cache[i], v)
			}
		}

		var b []byte
		var err error
		for idx := 0; idx < len(cache); idx++ {
			switch len(cache[idx]) {
			case 0, 1:
				continue
			case 2:
				b, err = sjson.SetBytes(b, cache[idx][0].String(), cache[idx][1].Value())
				if err != nil {
					return nil, fmt.Errorf("transform: array_zip: %v", err)
				}
			default:
				var vals []interface{}

				for i, v := range cache[idx] {
					if i == 0 {
						continue
					}

					vals = append(vals, v.Value())
				}

				b, err = sjson.SetBytes(b, cache[idx][0].String(), vals)
				if err != nil {
					return nil, fmt.Errorf("transform: array_zip: %v", err)
				}
			}
		}

		if err := msg.SetValue(tf.conf.Object.DstKey, b); err != nil {
			return nil, fmt.Errorf("transform: array_zip: %v", err)
		}

	case false:
		cache := make(map[int][]interface{})
		for _, val := range value.Array() {
			for i, v := range val.Array() {
				cache[i] = append(cache[i], v.Value())
			}
		}

		var b []interface{}
		for i := 0; i < len(cache); i++ {
			b = append(b, cache[i])
		}

		if err := msg.SetValue(tf.conf.Object.DstKey, b); err != nil {
			return nil, fmt.Errorf("transform: array_zip: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}

func (tf *arrayZip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
