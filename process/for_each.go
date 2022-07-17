package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/json"
)

/*
ForEachOptions contains custom options for the ForEach processor:
	Processor:
		processor to apply to the data
*/
type ForEachOptions struct {
	Processor config.Config
}

/*
ForEach processes data by iterating and applying a processor to each element in a JSON array. The processor supports these patterns:
	json:
		{"input":["ABC","DEF"]} >>> {"input":["ABC","DEF"],"output":["abc","def"]}

The processor uses this Jsonnet configuration:
	{
		type: 'for_each',
		settings: {
			input_key: 'input',
			output_key: 'output.-1',
			options: {
				processor: {
					type: 'case',
					settings: {
						options: {
							case: 'lower',
						}
					}
				},
			},
		},
	}
*/
type ForEach struct {
	Condition condition.OperatorConfig `json:"condition"`
	Options   ForEachOptions           `json:"options"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
}

// Slice processes a slice of bytes with the ForEach processor. Conditions are optionally applied on the bytes to enable processing.
func (p ForEach) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %v: %v", p, err)
	}

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %v: %v", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("slicer: %v", err)
		}
		slice = append(slice, processed)
	}

	return slice, nil
}

/*
Byte processes bytes with the ForEach processor.

Data is processed by iterating an input JSON array,
encapsulating the elements in a temporary JSON
object, and running the configured processor. This
technique avoids parsing errors when handling arrays
that contain JSON objects, such as:
	{"for_each":[{"foo":"bar"},{"foo":"baz"}]}

The temporary JSON object uses the configured
processor's name as its key (e.g., "case"). If the
configured processor has keys set (e.g., "foo"), then
the keys are concatenated (e.g., "case.foo"). The example
above produces this temporary JSON during processing:
	{"case":{"foo":"bar"}}
	{"case":{"foo":"baz"}}
*/
func (p ForEach) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// only supports json, error early if there are no keys
	if p.InputKey == "" && p.OutputKey == "" {
		return nil, fmt.Errorf("byter settings %v: %v", p, ProcessorInvalidSettings)
	}

	if _, ok := p.Options.Processor.Settings["input_key"]; ok {
		p.Options.Processor.Settings["input_key"] = p.Options.Processor.Type + "." + p.Options.Processor.Settings["input_key"].(string)
	} else {
		p.Options.Processor.Settings["input_key"] = p.Options.Processor.Type
	}

	if _, ok := p.Options.Processor.Settings["output_key"]; ok {
		p.Options.Processor.Settings["output_key"] = p.Options.Processor.Type + "." + p.Options.Processor.Settings["output_key"].(string)
	} else {
		p.Options.Processor.Settings["output_key"] = p.Options.Processor.Type
	}

	byter, err := ByterFactory(p.Options.Processor)
	if err != nil {
		return nil, err
	}

	value := json.Get(data, p.InputKey)
	if !value.IsArray() {
		return data, nil
	}

	for _, v := range value.Array() {
		var tmp []byte
		tmp, err := json.Set(tmp, p.Options.Processor.Type, v)
		if err != nil {
			return nil, fmt.Errorf("byter settings %v: %v", p, err)
		}

		tmp, err = byter.Byte(ctx, tmp)
		if err != nil {
			return nil, fmt.Errorf("byter settings %v: %v", p, err)
		}

		res := json.Get(tmp, p.Options.Processor.Type)
		data, err = json.Set(data, p.OutputKey, res)
		if err != nil {
			return nil, fmt.Errorf("byter settings %v: %v", p, err)
		}
	}

	return data, nil
}
