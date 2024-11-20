package condition

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
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

func (insp *stringEqualTo) Condition(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.HasFlag(message.IsControl) {
		return false, nil
	}

	compare := insp.b

	if insp.conf.Object.SourceKey == "" {
		return bytes.Equal(msg.Data(), compare), nil
	}

	val := msg.GetValue(insp.conf.Object.SourceKey)

	// for gjson's GetValue, if the path is empty string (indicating source key or target key is not present),
	// the Result.Exists() will return false
	// If source or target key is present but value cannot be found, always return false
	if !val.Exists() {
		return false, nil
	}

	target := msg.GetValue(insp.conf.Object.TargetKey)
	if target.Exists() {
		compare = target.Bytes()
	}

	return bytes.Equal(val.Bytes(), compare), nil
}

func (c *stringEqualTo) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
