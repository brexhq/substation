package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procErrConfig struct {
	// Error is the error message to return.
	Error string `json:"error"`
}

type procErr struct {
	conf procErrConfig
}

func newProcErr(_ context.Context, cfg config.Config) (*procErr, error) {
	conf := procErrConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	proc := procErr{
		conf: conf,
	}

	return &proc, nil
}

func (t *procErr) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procErr) Close(context.Context) error {
	return nil
}

func (t *procErr) Transform(_ context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	return messages, fmt.Errorf("%s", t.conf.Error)
}
