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

func (t *procDelete) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procDelete) Close(context.Context) error {
	return nil
}

func (t *procDelete) Transform(_ context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		if err := message.Delete(t.conf.Key); err != nil {
			return nil, fmt.Errorf("transform: proc_delete: %v", err)
		}

		output = append(output, message)
	}

	return output, nil
}
