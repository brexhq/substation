package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type utilErrConfig struct {
	// Error is the error message to return.
	Error string `json:"error"`
}

type utilErr struct {
	conf utilErrConfig
}

func newUtilErr(_ context.Context, cfg config.Config) (*utilErr, error) {
	conf := utilErrConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	proc := utilErr{
		conf: conf,
	}

	return &proc, nil
}

func (util *utilErr) String() string {
	b, _ := gojson.Marshal(util.conf)
	return string(b)
}

func (*utilErr) Close(context.Context) error {
	return nil
}

func (util *utilErr) Transform(_ context.Context, message *mess.Message) ([]*mess.Message, error) {
	return []*mess.Message{message}, fmt.Errorf("%s", util.conf.Error)
}
