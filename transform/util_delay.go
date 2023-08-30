package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
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
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_util_delay: %v", err)
	}

	// Validate required options.
	if conf.Duration == "" {
		return nil, fmt.Errorf("transform: new_util_delay: duration: %v", errors.ErrMissingRequiredOption)
	}

	dur, err := time.ParseDuration(conf.Duration)
	if err != nil {
		return nil, fmt.Errorf("transform: new_util_delay: duration: %v", err)
	}

	tf := utilDelay{
		conf: conf,
		dur:  dur,
	}

	return &tf, nil
}

func (tf *utilDelay) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*utilDelay) Close(context.Context) error {
	return nil
}

func (tf *utilDelay) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	time.Sleep(tf.dur)
	return []*message.Message{msg}, nil
}
