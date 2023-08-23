package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type procDeleteConfig struct {
	// Key is the object key to delete.
	Key string `json:"key"`
}

type procDelete struct {
	conf procDeleteConfig
}

func newProcDelete(_ context.Context, cfg config.Config) (*procDelete, error) {
	conf := procDeleteConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	if conf.Key == "" {
		return nil, fmt.Errorf("transform: proc_delete: key %q: %v", conf.Key, errInvalidDataPattern)
	}

	proc := procDelete{
		conf: conf,
	}

	return &proc, nil
}

func (proc *procDelete) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procDelete) Close(context.Context) error {
	return nil
}

func (proc *procDelete) Transform(_ context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	if err := message.Delete(proc.conf.Key); err != nil {
		return nil, fmt.Errorf("transform: proc_delete: %v", err)
	}

	return []*mess.Message{message}, nil
}
