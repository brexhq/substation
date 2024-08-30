package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newStringStartsWith(_ context.Context, cfg config.Config) (*stringStartsWith, error) {
	conf := stringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := stringStartsWith{
		conf: conf,
		b:    []byte(conf.Value),
	}

	return &insp, nil
}

type stringStartsWith struct {
	conf stringConfig

	b []byte
}

func (insp *stringStartsWith) Condition(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		return bytes.HasPrefix(msg.Data(), insp.b), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return bytes.HasPrefix(value.Bytes(), insp.b), nil
}

func (c *stringStartsWith) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
