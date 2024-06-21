package condition

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNumberEqualTo(_ context.Context, cfg config.Config) (*numberEqualTo, error) {
	conf := numberConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}
	insp := numberEqualTo{
		conf: conf,
	}
	return &insp, nil
}

type numberEqualTo struct {
	conf numberConfig
}

func (insp *numberEqualTo) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
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

func (c *numberEqualTo) match(f float64) bool {
	return f == c.conf.Value
}

func (c *numberEqualTo) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
