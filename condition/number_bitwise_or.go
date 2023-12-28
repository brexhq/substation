package condition

import (
	"context"
	"strconv"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newNumberBitwiseOR(_ context.Context, cfg config.Config) (*numberBitwiseOR, error) {
	conf := numberBitwiseConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := numberBitwiseOR{
		conf: conf,
	}

	return &insp, nil
}

type numberBitwiseOR struct {
	conf numberBitwiseConfig
}

func (insp *numberBitwiseOR) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SrcKey == "" {
		value, err := strconv.ParseInt(string(msg.Data()), 10, 64)
		if err != nil {
			return false, err
		}

		return value|insp.conf.Value != 0, nil
	}

	value := msg.GetValue(insp.conf.Object.SrcKey)
	return value.Int()|insp.conf.Value != 0, nil
}
