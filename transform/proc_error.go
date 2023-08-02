package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procErrorConfig struct {
	// Error is the error message to return.
	Error string `json:"error"`
}

type procError struct {
	conf procErrorConfig
}

func newProcError(_ context.Context, cfg config.Config) (*procError, error) {
	conf := procErrorConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	proc := procError{
		conf: conf,
	}

	return &proc, nil
}

func (t *procError) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procError) Close(context.Context) error {
	return nil
}

func (t *procError) Transform(_ context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	return messages, fmt.Errorf("%s", t.conf.Error)
}
