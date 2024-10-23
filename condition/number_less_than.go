package condition

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNumberLessThan(_ context.Context, cfg config.Config) (*numberLessThan, error) {
	conf := numberConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}
	insp := numberLessThan{
		conf: conf,
	}
	return &insp, nil
}

type numberLessThan struct {
	conf numberConfig
}

func (insp *numberLessThan) Condition(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.HasFlag(message.IsControl) {
		return false, nil
	}

	compare := insp.conf.Value

	if insp.conf.Object.SourceKey == "" {
		f, err := strconv.ParseFloat(string(msg.Data()), 64)
		if err != nil {
			return false, err
		}

		return insp.match(f, compare), nil
	}

	target := msg.GetValue(insp.conf.Object.TargetKey)

	if target.Exists() {
		compare = target.Float()
	}

	v := msg.GetValue(insp.conf.Object.SourceKey)
	return insp.match(v.Float(), compare), nil
}

func (c *numberLessThan) match(f float64, t float64) bool {
	return f < t
}

func (c *numberLessThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
