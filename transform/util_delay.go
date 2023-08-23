package transform

import (
	"context"
	gojson "encoding/json"
	"time"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type utilDelayConfig struct {
	Duration string `json:"duration"`
}

type utilDelay struct {
	conf utilDelayConfig

	dur time.Duration
}

func newUtilDelay(_ context.Context, cfg config.Config) (*utilDelay, error) {
	conf := utilDelayConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	dur, err := time.ParseDuration(conf.Duration)
	if err != nil {
		return nil, err
	}

	util := utilDelay{
		conf: conf,
		dur:  dur,
	}

	return &util, nil
}

func (util *utilDelay) String() string {
	b, _ := gojson.Marshal(util.conf)
	return string(b)
}

func (*utilDelay) Close(context.Context) error {
	return nil
}

func (util *utilDelay) Transform(_ context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	time.Sleep(util.dur)
	return []*mess.Message{message}, nil
}
