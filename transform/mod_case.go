package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modCaseConfig struct {
	Object configObject `json:"object"`

	// Type is the case formatting that is applied.
	//
	// Must be one of:
	//	- upcase
	//	- downcase
	//	- snakecase
	Type string `json:"type"`
}

type modCase struct {
	conf  modCaseConfig
	isObj bool
}

func newModCase(_ context.Context, cfg config.Config) (*modCase, error) {
	conf := modCaseConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_case: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_case: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_case: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: new_mod_case: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"upcase",
			"upper",
			"downcase",
			"lower",
			"snakecase",
			"snake",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: new_mod_case: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	tf := modCase{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *modCase) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modCase) Close(context.Context) error {
	return nil
}

func (tf *modCase) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObj {
		var value []byte
		switch tf.conf.Type {
		case "upcase", "upper":
			value = bytes.ToUpper(msg.Data())
		case "downcase", "lower":
			value = bytes.ToLower(msg.Data())
		case "snakecase", "snake":
			value = []byte(strcase.ToSnake(string(msg.Data())))
		}

		finMsg := message.New().SetData(value).SetMetadata(msg.Metadata())
		return []*message.Message{finMsg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key).String()

	var value string
	switch tf.conf.Type {
	case "upcase", "upper":
		value = strings.ToUpper(result)
	case "downcase", "lower":
		value = strings.ToLower(result)
	case "snakecase", "snake":
		value = strcase.ToSnake(result)
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_case: %v", err)
	}

	return []*message.Message{msg}, nil
}
