package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newStringEndsWith(_ context.Context, cfg config.Config) (*stringEndsWith, error) {
	conf := stringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := stringEndsWith{
		conf: conf,
		b:    []byte(conf.Value),
	}

	return &insp, nil
}

type stringEndsWith struct {
	conf stringConfig

	b []byte
}

func (insp *stringEndsWith) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		return bytes.HasSuffix(msg.Data(), insp.b), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return bytes.HasSuffix(value.Bytes(), insp.b), nil
}

func (c *stringEndsWith) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
