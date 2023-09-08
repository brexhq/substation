package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type strCaseUpConfig struct {
	Object iconfig.Object `json:"object"`
}

func (c *strCaseUpConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *strCaseUpConfig) Validate() error {
	if c.Object.Key == "" && c.Object.SetKey != "" {
		return fmt.Errorf("object_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.Key != "" && c.Object.SetKey == "" {
		return fmt.Errorf("object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type strCaseUp struct {
	conf     strCaseUpConfig
	isObject bool
}

func newStrCaseUp(_ context.Context, cfg config.Config) (*strCaseUp, error) {
	conf := strCaseUpConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("strings: new_case_up: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("strings: new_case_up: %v", err)
	}

	tf := strCaseUp{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *strCaseUp) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		b := bytes.ToUpper(msg.Data())
		msg.SetData(b)

		return []*message.Message{msg}, nil
	}

	v := msg.GetValue(tf.conf.Object.Key)
	s := strings.ToUpper(v.String())

	if err := msg.SetValue(tf.conf.Object.SetKey, s); err != nil {
		return nil, fmt.Errorf("transform: strings_case_up: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *strCaseUp) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

func (*strCaseUp) Close(context.Context) error {
	return nil
}
