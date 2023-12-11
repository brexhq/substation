package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newStringEqualTo(_ context.Context, cfg config.Config) (*stringEqualTo, error) {
	conf := stringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := stringEqualTo{
		conf: conf,
		b:    []byte(conf.Value),
	}

	return &insp, nil
}

type stringEqualTo struct {
	conf stringConfig

	b []byte
}

func (insp *stringEqualTo) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		return bytes.Equal(msg.Data(), insp.b), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	return bytes.Equal(value.Bytes(), insp.b), nil
}

func (c *stringEqualTo) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
