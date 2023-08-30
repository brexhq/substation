package transform

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type sendStdoutConfig struct{}

type sendStdout struct {
	conf sendStdoutConfig
}

func newSendStdout(_ context.Context, cfg config.Config) (*sendStdout, error) {
	conf := sendStdoutConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_send_stdout: %v", err)
	}

	tf := sendStdout{
		conf: conf,
	}

	return &tf, nil
}

func (*sendStdout) Close(context.Context) error {
	return nil
}

func (*sendStdout) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	fmt.Println(string(msg.Data()))
	return []*message.Message{msg}, nil
}
