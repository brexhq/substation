package process

/*

 */

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// used with json.Get, returns a pretty printed root JSON object
const ppModifier = `@this|@pretty`
const openCurlyBracket = 123  // {
const closeCurlyBracket = 125 // }

/*
PrettyPrintUnbalancedBrackets is returned when the processor is given input
that does not contain an equal number of open curly brackets ( { ) and close
curly brackets ( } ). The most common causes of this error are invalid input JSON
(e.g., `{{"foo":"bar"}`) or using the processor with multi-core processing enabled.
*/
const PrettyPrintUnbalancedBrackets = errors.Error("PrettyPrintUnbalancedBrackets")

/*
PrettyPrintOptions contains custom options settings for the PrettyPrint processor:
	Direction:
		the direction of the pretty transformation
		must be one of:
			to: applies prettyprint formatting
			from: reverses prettyprint formatting
*/
type PrettyPrintOptions struct {
	Direction string `json:"direction"`
}

/*
PrettyPrint processes data by applying or reversing prettyprint formatting to JSON.
This processor has significant limitations when used to reverse prettyprint, including:
	- cannot support multi-core processing
	- invalid input will cause unpredictable results

It is strongly recommended to _not_ use this processor unless absolutely necessary; a
more reliable solution is to modify the source application emitting the multi-line JSON
object so that it outputs a single-line object instead.

The processor supports these patterns:
	data:
		{
			"foo": "bar"
		}  >>> {"foo":"bar"}

		{"foo":"bar"} >>> {
			"foo": "bar"
		}

The processor uses this Jsonnet configuration:
	{
		type: 'pretty_print',
		settings: {
			options: {
				direction: 'from',
			},
		},
	}
*/
type PrettyPrint struct {
	Options   PrettyPrintOptions       `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
}

/*
Slice processes bytes with the PrettyPrint processor.

Applying prettyprint formatting is handled by the
gjson PrettyPrint modifier and is applied to the root
JSON object.

Reversing prettyprint formatting is handled by
iterating incoming data per byte and pushing the
bytes to a stack. When an equal number of open
and close curly brackets ( { } ) are observed,
then the stack of bytes has JSON compaction
applied and the result is emitted as a new object.
*/
func (p PrettyPrint) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
	}

	var count int
	var stack []byte

	slice := NewSlice(&s)
	for _, data := range s {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, fmt.Errorf("slicer settings %+v: %w", p, err)
		}

		if !ok {
			slice = append(slice, data)
			continue
		}

		switch p.Options.Direction {
		case "to":
			s := json.Get(data, ppModifier).String()
			slice = append(slice, []byte(s))

		case "from":
			for _, d := range data {
				stack = append(stack, d)

				if d == openCurlyBracket {
					count++
				}

				if d == closeCurlyBracket {
					count--
				}

				if count == 0 {
					var buf bytes.Buffer
					if err := gojson.Compact(&buf, stack); err != nil {
						return nil, err
					}

					slice = append(slice, buf.Bytes())
					stack = []byte{}
				}
			}

		default:
			return nil, fmt.Errorf("slicer settings %+v: %w", p, ProcessorInvalidSettings)
		}
	}

	if count != 0 {
		return nil, fmt.Errorf("slicer settings %+v: %w", p, PrettyPrintUnbalancedBrackets)
	}

	return slice, nil
}

/*
Byte processes bytes with the PrettyPrint processor.

Applying prettyprint formatting is handled by the
gjson PrettyPrint modifier and is applied to the root
JSON object.

Byte _does not_ support reversing prettyprint formatting;
this support is unnecessary for multi-line JSON objects
that are stored in a single byte array.
*/
func (p PrettyPrint) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return nil, fmt.Errorf("byter settings %+v: %w", p, ProcessorInvalidSettings)
	}

	if p.Options.Direction == "to" {
		value := json.Get(data, ppModifier).String()
		return []byte(value), nil
	}

	return nil, fmt.Errorf("slicer settings %+v: %w", p, ProcessorInvalidSettings)
}
