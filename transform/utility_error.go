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
	Error string `json:"error"`
}

func (c *utilErrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityError(_ context.Context, cfg config.Config) (*utilityError, error) {
	conf := utilErrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_util_error: %v", err)
	}

	tf := utilityError{
		conf: conf,
	}

	return &tf, nil
}

type utilityError struct {
	conf utilErrConfig
}

func (tf *utilityError) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	return []*message.Message{msg}, fmt.Errorf("%s", tf.conf.Error)
}

func (tf *utilityError) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*utilityError) Close(context.Context) error {
	return nil
}
