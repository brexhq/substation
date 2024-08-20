package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/v2/config"
)

func newStringContains(_ context.Context, cfg config.Config) (*stringContains, error) {
	conf := stringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := stringContains{
		conf: conf,
		b:    []byte(conf.Value),
	}

	return &insp, nil
}

type stringContains struct {
	conf stringConfig

	b []byte
}

func (insp *stringContains) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		return bytes.Contains(msg.Data(), insp.b), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return bytes.Contains(value.Bytes(), insp.b), nil
}

func (c *stringContains) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
