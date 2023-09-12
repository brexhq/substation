package condition

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

type logicLenLessThan struct {
	conf logicLengthConfig
}

func newLogicLenLessThan(_ context.Context, cfg config.Config) (*logicLenLessThan, error) {
	conf := logicLengthConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := logicLenLessThan{
		conf: conf,
	}

	return &insp, nil
}

func (insp *logicLenLessThan) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		llm := logicLengthMeasurement(msg.Data(), insp.conf.Measurement)
		return insp.match(llm), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	if value.IsArray() {
		l := len(value.Array())
		return insp.match(l), nil
	}

	llm := logicLengthMeasurement(value.Bytes(), insp.conf.Measurement)
	return insp.match(llm), nil
}

func (c *logicLenLessThan) match(length int) bool {
	return length < c.conf.Length
}

func (c *logicLenLessThan) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
