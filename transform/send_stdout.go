package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/message"
)

type sendStdoutConfig struct{}

func (c *sendStdoutConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

type sendStdout struct {
	conf sendStdoutConfig
}

func newSendStdout(_ context.Context, cfg config.Config) (*sendStdout, error) {
	conf := sendStdoutConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_send_stdout: %v", err)
	}

	tf := sendStdout{
		conf: conf,
	}

	return &tf, nil
}

func (*sendStdout) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	fmt.Println(string(msg.Data()))
	return []*message.Message{msg}, nil
}

func (tf *sendStdout) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*sendStdout) Close(context.Context) error {
	return nil
}
