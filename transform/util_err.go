package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type utilErrConfig struct {
	Error string `json:"error"`
}

type utilErr struct {
	conf utilErrConfig
}

func newUtilErr(_ context.Context, cfg config.Config) (*utilErr, error) {
	conf := utilErrConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_util_err: %v", err)
	}

	tf := utilErr{
		conf: conf,
	}

	return &tf, nil
}

func (tf *utilErr) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*utilErr) Close(context.Context) error {
	return nil
}

func (tf *utilErr) Transform(_ context.Context, msg *message.Message) ([]*message.Message, error) {
	return []*message.Message{msg}, fmt.Errorf("%s", tf.conf.Error)
}
