package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

type sendStdoutConfig struct{}

type sendStdout struct {
	conf sendStdoutConfig
}

// Create a new stdout send.
func newSendStdout(_ context.Context, cfg config.Config) (*sendStdout, error) {
	conf := sendStdoutConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
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
