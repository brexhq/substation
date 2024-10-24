package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newStringLessThan(_ context.Context, cfg config.Config) (*stringLessThan, error) {
	conf := stringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := stringLessThan{
		conf: conf,
		b:    []byte(conf.Value),
	}

	return &insp, nil
}

type stringLessThan struct {
	conf stringConfig

	b []byte
}

func (insp *stringLessThan) Condition(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.HasFlag(message.IsControl) {
		return false, nil
	}

	compare := insp.b

	if insp.conf.Object.SourceKey == "" {
		return bytes.Compare(msg.Data(), compare) < 0, nil
	}
	target := msg.GetValue(insp.conf.Object.TargetKey)

	if target.Exists() {
		compare = target.Bytes()
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return bytes.Compare(value.Bytes(), compare) < 0, nil
}

func (c *stringLessThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
