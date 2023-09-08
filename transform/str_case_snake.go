package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/iancoleman/strcase"
)

type strCaseSnakeConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *strCaseSnakeConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strCaseSnakeConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type strCaseSnake struct {
	conf     strCaseSnakeConfig
	isObject bool
}

func newStrCaseSnake(_ context.Context, cfg config.Config) (*strCaseSnake, error) {
	conf := strCaseSnakeConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_snake: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_str_case_snake: %v", err)
	}

	tf := strCaseSnake{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *strCaseSnake) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := []byte(strcase.ToSnake(string(msg.Data())))
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key)
	s := strcase.ToSnake(v.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, s); err != nil {
		return nil, fmt.Errorf("transform: str_case_snake: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *strCaseSnake) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*strCaseSnake) Close(context.Context) error {
	return nil
}
