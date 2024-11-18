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
	source_value := msg.GetValue(insp.conf.Object.SourceKey)
	target_value := msg.GetValue(insp.conf.Object.TargetKey)

	// for gjson's GetValue, if the path is empty string (indicating source key or target key is not present),
	// the Result.Exists() will return false
	// If source or target key is present but value cannot be found, always return false
	if !source_value.Exists() || insp.conf.Object.TargetKey != "" && !target_value.Exists() {
		return false, nil
	}

	if target_value.Exists() {
		compare = target_value.Float()
	}

	return insp.match(source_value.Float(), compare), nil
}

func (c *numberEqualTo) match(f float64, t float64) bool {
	return f == t
}

func (c *numberEqualTo) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
