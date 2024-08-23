package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
	"github.com/brexhq/substation/v2/message"
)

type utilityDelayConfig struct {
	// Duration is the amount of time to delay.
	Duration string `json:"duration"`

	ID string `json:"id"`
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

func newUtilityDelay(_ context.Context, cfg config.Config) (*utilityDelay, error) {
	conf := utilityDelayConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform utility_delay: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "utility_delay"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	dur, err := time.ParseDuration(conf.Duration)
	if err != nil {
		return nil, fmt.Errorf("transform %s: duration: %v", conf.ID, err)
	}

	tf := utilityDelay{
		conf: conf,
		dur:  dur,
	}

	return &tf, nil
}

type utilityDelay struct {
	conf utilityDelayConfig

	dur time.Duration
}

func (tf *utilityDelay) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	time.Sleep(tf.dur)
	return []*message.Message{msg}, nil
}

func (tf *utilityDelay) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
