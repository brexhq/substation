package condition

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type numberLengthEqualTo struct {
	conf numberLengthConfig
}

func newNumberLengthEqualTo(_ context.Context, cfg config.Config) (*numberLengthEqualTo, error) {
	conf := numberLengthConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := numberLengthEqualTo{
		conf: conf,
	}

	return &insp, nil
}

func (insp *numberLengthEqualTo) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		llm := numberLengthMeasurement(msg.Data(), insp.conf.Measurement)
		return insp.match(llm), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	if value.IsArray() {
		l := len(value.Array())
		return insp.match(l), nil
	}

	llm := numberLengthMeasurement(value.Bytes(), insp.conf.Measurement)
	return insp.match(llm), nil
}

func (c *numberLengthEqualTo) match(length int) bool {
	return length == c.conf.Length
}

func (c *numberLengthEqualTo) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
