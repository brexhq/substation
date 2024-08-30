package condition

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNumberLengthGreaterThan(_ context.Context, cfg config.Config) (*numberLengthGreaterThan, error) {
	conf := numberLengthConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := numberLengthGreaterThan{
		conf: conf,
	}

	return &insp, nil
}

type numberLengthGreaterThan struct {
	conf numberLengthConfig
}

func (insp *numberLengthGreaterThan) Condition(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		llm := numberLengthMeasurement(msg.Data(), insp.conf.Measurement)
		return insp.match(llm), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	if value.IsArray() {
		l := len(value.Array())
		return insp.match(l), nil
	}

	llm := numberLengthMeasurement(value.Bytes(), insp.conf.Measurement)
	return insp.match(llm), nil
}

func (c *numberLengthGreaterThan) match(length int) bool {
	return length > c.conf.Value
}

func (c *numberLengthGreaterThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
