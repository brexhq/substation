package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type sendStdoutConfig struct{}

type sendStdout struct {
	conf sendStdoutConfig
}

func newSendStdout(_ context.Context, cfg config.Config) (*sendStdout, error) {
	conf := sendStdoutConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	send := sendStdout{
		conf: conf,
	}

	return &send, nil
}

func (*sendStdout) Close(context.Context) error {
	return nil
}

func (*sendStdout) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	fmt.Println(string(message.Data()))
	return []*mess.Message{message}, nil
}
