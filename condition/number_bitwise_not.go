package condition

import (
	"context"
	"strconv"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

func newNumberBitwiseNOT(_ context.Context, cfg config.Config) (*numberBitwiseNOT, error) {
	conf := numberBitwiseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := numberBitwiseNOT{
		conf: conf,
	}

	return &insp, nil
}

type numberBitwiseNOT struct {
	conf numberBitwiseConfig
}

func (insp *numberBitwiseNOT) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		value, err := strconv.ParseInt(string(msg.Data()), 10, 64)
		if err != nil {
			return false, err
		}

		return ^value != 0, nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return ^value.Int() != 0, nil
}
