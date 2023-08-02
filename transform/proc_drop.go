package transform

import (
	"context"
	gojson "encoding/json"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type procDropConfig struct{}

type procDrop struct {
	conf procDropConfig
}

func newProcDrop(_ context.Context, cfg config.Config) (*procDrop, error) {
	conf := procDropConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
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

func (t *procDrop) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		if message.IsControl() {
			output = append(output, message)
		}
	}

	return output, nil
}
