package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type utilErrConfig struct {
	Message string `json:"message"`
}

func (c *utilErrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityErr(_ context.Context, cfg config.Config) (*utilityErr, error) {
	conf := utilErrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_util_error: %v", err)
	}

	tf := utilityErr{
		conf: conf,
	}

	return &tf, nil
}

type utilityErr struct {
	conf utilErrConfig
}

func (tf *utilityErr) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	return []*message.Message{msg}, fmt.Errorf("%s", tf.conf.Message)
}

func (tf *utilityErr) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}