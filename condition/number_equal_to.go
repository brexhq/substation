package condition

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
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

func (insp *numberEqualTo) Condition(ctx context.Context, msg *message.Message) (output bool, err error) {
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


	val := msg.GetValue(insp.conf.Object.SourceKey)

	// for gjson's GetValue, if the path is empty string (indicating source key or target key is not present),
	// the Result.Exists() will return false
	// If source or target key is present but value cannot be found, always return false
	if !val.Exists() {
		return false, nil
	}

	target := msg.GetValue(insp.conf.Object.TargetKey)
	if target.Exists() {
		compare = target.Float()
	}

	return insp.match(val.Float(), compare), nil
}

func (c *numberEqualTo) match(f float64, t float64) bool {
	return f == t
}

func (c *numberEqualTo) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
