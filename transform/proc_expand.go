package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
	"github.com/tidwall/gjson"
)

type procExpandConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
}

type procExpand struct {
	conf procExpandConfig
}

func newProcExpand(_ context.Context, cfg config.Config) (*procExpand, error) {
	conf := procExpandConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	proc := procExpand{
		conf: conf,
	}

	return &proc, nil
}

func (proc *procExpand) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procExpand) Close(context.Context) error {
	return nil
}

func (proc *procExpand) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	var output []*mess.Message

	// Data is processed by retrieving and iterating an
	// array containing JSON objects (result) and setting
	// any remaining keys from the object (remains) into
	// each new object. If there is no Key, then the input
	// is treated as an array.
	//
	// input:
	// 	{"expand":[{"foo":"bar"},{"baz":"qux"}],"quux":"corge"}
	// result:
	//  [{"foo":"bar"},{"baz":"qux"}]
	// remains:
	// 	{"quux":"corge"}
	// output:
	// 	{"foo":"bar","quux":"corge"}
	// 	{"baz":"qux","quux":"corge"}
	var result, remains gjson.Result

	if proc.conf.Key != "" {
		result = json.Get(message.Data(), proc.conf.Key)

		// Deleting the key from the object speeds
		// up processing large objects.
		if err := message.Delete(proc.conf.Key); err != nil {
			return nil, fmt.Errorf("transform: proc_expand: %v", err)
		}

		remains = json.Get(message.Data(), "@this")
	} else {
		// remains is unused when there is no key
		result = json.Get(message.Data(), "@this")
	}

	for _, res := range result.Array() {
		// Data processing. Elements from the array become new data.
		if proc.conf.Key == "" {
			msg, err := mess.New(
				mess.SetData([]byte(res.String())),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_expand: %v", err)
			}

			output = append(output, msg)
			continue
		}

		msg, err := mess.New(
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_expand: %v", err)
		}

		// Object processing. Remaining keys from the original object are added
		// to the new object.
		for key, val := range remains.Map() {
			if err := msg.Set(key, val); err != nil {
				return nil, fmt.Errorf("transform: proc_expand: %v", err)
			}
		}

		if proc.conf.SetKey != "" {
			if err := msg.Set(proc.conf.SetKey, res); err != nil {
				return nil, fmt.Errorf("transform: proc_expand: %v", err)
			}

			output = append(output, msg)
			continue
		}

		// At this point there should be two objects that need to be
		// merged into a single object. the objects are merged using
		// the GJSON @join function, which joins all objects that are
		// in an array. if the array contains non-object data, then
		// it is ignored.
		//
		// [{"foo":"bar"},{"baz":"qux"}}] becomes {"foo":"bar","baz":"qux"}
		// [{"foo":"bar"},{"baz":"qux"},"quux"] becomes {"foo":"bar","baz":"qux"}
		tmp := fmt.Sprintf(`[%s,%s]`, msg.Data(), res.String())
		join := json.Get([]byte(tmp), "@join")

		msg, err = mess.New(
			mess.SetData([]byte(join.String())),
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_expand: %v", err)
		}

		output = append(output, msg)
	}

	return output, nil
}
