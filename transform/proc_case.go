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
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type procCaseConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Type is the case formatting that is applied.
	//
	// Must be one of:
	//
	// - upper
	//
	// - lower
	//
	// - snake
	Type string `json:"type"`
}

type procCase struct {
	conf     procCaseConfig
	isObject bool
}

func newProcCase(_ context.Context, cfg config.Config) (*procCase, error) {
	conf := procCaseConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") || (conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_case: %v", errInvalidDataPattern)
	}

	if conf.Type == "" {
		return nil, fmt.Errorf("transform: proc_case: type: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{
			"upper",
			"lower",
			"snake",
		},
		conf.Type) {
		return nil, fmt.Errorf("transform: proc_case: type %q: %v", conf.Type, errors.ErrInvalidOption)
	}

	proc := procCase{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	return &proc, nil
}

func (t *procCase) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procCase) Close(context.Context) error {
	return nil
}

func (t *procCase) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message
	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key).String()

			var value string
			switch t.conf.Type {
			case "upper":
				value = strings.ToUpper(result)
			case "lower":
				value = strings.ToLower(result)
			case "snake":
				value = strcase.ToSnake(result)
			}

			if err := message.Set(t.conf.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: proc_case: %v", err)
			}

			output = append(output, message)
		case false:
			var value []byte
			switch t.conf.Type {
			case "upper":
				value = bytes.ToUpper(message.Data())
			case "lower":
				value = bytes.ToLower(message.Data())
			case "snake":
				value = []byte(strcase.ToSnake(string(message.Data())))
			}

			msg, err := mess.New(
				mess.SetData(value),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_case: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}
