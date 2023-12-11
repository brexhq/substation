package condition

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNumberLengthLessThan(_ context.Context, cfg config.Config) (*numberLengthLessThan, error) {
	conf := numberLengthConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := numberLengthLessThan{
		conf: conf,
	}

	return &insp, nil
}

type numberLengthLessThan struct {
	conf numberLengthConfig
}

func (insp *numberLengthLessThan) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
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

func (c *numberLengthLessThan) match(length int) bool {
	return length < c.conf.Value
}

func (c *numberLengthLessThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
