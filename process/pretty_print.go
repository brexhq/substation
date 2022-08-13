package process

/*

 */

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
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
ApplyBatch processes a slice of encapsulated data
with the PrettyPrint processor.

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
func (p PrettyPrint) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return nil, fmt.Errorf("unary settings %+v: %w", p, ProcessorInvalidSettings)
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("batch settings %+v: %w", p, err)
	}

	var count int
	var stack []byte

	slice := NewBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(cap)
		if err != nil {
			return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
		}

		if !ok {
			slice = append(slice, cap)
			continue
		}

		switch p.Options.Direction {
		case "to":
			s := cap.Get(ppModifier).String()
			cap.SetData([]byte(s))
			slice = append(slice, cap)

		case "from":
			for _, data := range cap.GetData() {
				stack = append(stack, data)

				if data == openCurlyBracket {
					count++
				}

				if data == closeCurlyBracket {
					count--
				}

				if count == 0 {
					var buf bytes.Buffer
					if err := gojson.Compact(&buf, stack); err != nil {
						return nil, fmt.Errorf("applybatch settings %+v: %w", p, err)
					}

					if json.Valid(buf.Bytes()) {
						newCap := config.NewCapsule()
						newCap.SetData(buf.Bytes())
						slice = append(slice, newCap)
					}

					stack = []byte{}
				}
			}

		default:
			return nil, fmt.Errorf("applybatch settings %+v: %w", p, ProcessorInvalidSettings)
		}
	}

	if count != 0 {
		return nil, fmt.Errorf("applybatch settings %+v: %w", p, PrettyPrintUnbalancedBrackets)
	}

	return slice, nil
}

/*
Apply processes encapsulated data with the PrettyPrint
processor.

Applying prettyprint formatting is handled by the
gjson PrettyPrint modifier and is applied to the root
JSON object.

Byte _does not_ support reversing prettyprint formatting;
this support is unnecessary for multi-line JSON objects
that are stored in a single byte array.
*/
func (p PrettyPrint) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return cap, fmt.Errorf("unary settings %+v: %w", p, ProcessorInvalidSettings)
	}

	if p.Options.Direction == "to" {
		cap.SetData([]byte(cap.Get(ppModifier).String()))
		return cap, nil
	}

	return cap, fmt.Errorf("unary settings %+v: %w", p, ProcessorInvalidSettings)
}
