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

func (t *sendStdout) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	for _, message := range messages {
		if message.IsControl() {
			continue
		}

		fmt.Println(string(message.Data()))
	}

	return messages, nil
}
