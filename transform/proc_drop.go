package transform

import (
	"context"
	gojson "encoding/json"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type procDropConfig struct{}

type procDrop struct {
	conf procDropConfig
}

func newProcDrop(_ context.Context, cfg config.Config) (*procDrop, error) {
	conf := procDropConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	proc := procDrop{
		conf: conf,
	}

	return &proc, nil
}

func (t *procDrop) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procDrop) Close(context.Context) error {
	return nil
}

func (t *procDrop) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	return nil, nil
}
