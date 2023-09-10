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

type utilityDelayConfig struct {
	Duration string `json:"duration"`
}

func (c *utilityDelayConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *utilityDelayConfig) Validate() error {
	if c.Duration == "" {
		return fmt.Errorf("duration: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type utilityDelay struct {
	conf utilityDelayConfig

	dur time.Duration
}

func newUtilityDelay(_ context.Context, cfg config.Config) (*utilityDelay, error) {
	conf := utilityDelayConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_util_delay: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_util_delay: %v", err)
	}

	dur, err := time.ParseDuration(conf.Duration)
	if err != nil {
		return nil, fmt.Errorf("transform: new_util_delay: duration: %v", err)
	}

	tf := utilityDelay{
		conf: conf,
		dur:  dur,
	}

	return &tf, nil
}

func (tf *utilityDelay) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	time.Sleep(tf.dur)
	return []*message.Message{msg}, nil
}

func (tf *utilityDelay) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*utilityDelay) Close(context.Context) error {
	return nil
}
