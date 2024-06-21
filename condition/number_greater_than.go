package condition

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNumberGreaterThan(_ context.Context, cfg config.Config) (*numberGreaterThan, error) {
	conf := numberConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := numberGreaterThan{
		conf: conf,
	}

	return &insp, nil
}

type numberGreaterThan struct {
	conf numberConfig
}

func (insp *numberGreaterThan) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		f, err := strconv.ParseFloat(string(msg.Data()), 64)
		if err != nil {
			return false, err
		}

		return insp.match(f), nil
	}

	v := msg.GetValue(insp.conf.Object.SourceKey)
	return insp.match(v.Float()), nil
}

func (c *numberGreaterThan) match(f float64) bool {
	return f > c.conf.Value
}

func (c *numberGreaterThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
