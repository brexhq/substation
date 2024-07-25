package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newStringGreaterThan(_ context.Context, cfg config.Config) (*stringGreaterThan, error) {
	conf := stringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := stringGreaterThan{
		conf: conf,
		b:    []byte(conf.Value),
	}

	return &insp, nil
}

type stringGreaterThan struct {
	conf stringConfig

	b []byte
}

func (insp *stringGreaterThan) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	compare := insp.b

	target := msg.GetValue(insp.conf.Object.TargetKey)

	if target.Exists() {
		compare = target.Bytes()
	}

	if insp.conf.Object.SourceKey == "" {
		return bytes.Compare(msg.Data(), compare) > 0, nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return bytes.Compare(value.Bytes(), compare) > 0, nil
}

func (c *stringGreaterThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
