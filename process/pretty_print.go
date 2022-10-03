package process

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

const (
	// used with json.Get, returns a pretty printed root JSON object
	ppModifier          = `@this|@pretty`
	ppOpenCurlyBracket  = 123 // {
	ppCloseCurlyBracket = 125 // }
)

/*
errPrettyPrintIncompleteJSON is returned when the processor is given input
that does not contain an equal number of open curly brackets ( { ) and close
curly brackets ( } ), indicating that the input was an incomplete JSON object.
The most common causes of this error are invalid input JSON
(e.g., `{{"foo":"bar"}`) or using the processor with multi-core processing enabled.
*/
const errPrettyPrintIncompleteJSON = errors.Error("incomplete JSON object")

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

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "pretty_print",
		"settings": {
			"options": {
				"direction": "from"
			}
		}
	}
*/
type PrettyPrint struct {
	Options   PrettyPrintOptions `json:"options"`
	Condition condition.Config   `json:"condition"`
}

/*
PrettyPrintOptions contains custom options settings for the PrettyPrint processor:
	Direction:
		direction of the pretty transformation
		must be one of:
			to: applies prettyprint formatting
			from: reverses prettyprint formatting
*/
type PrettyPrintOptions struct {
	Direction string `json:"direction"`
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
		return nil, fmt.Errorf("process pretty_print: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("process pretty_print: %v", err)
	}

	var count int
	var stack []byte

	newCaps := newBatch(&caps)
	for _, cap := range caps {
		ok, err := op.Operate(ctx, cap)
		if err != nil {
			return nil, fmt.Errorf("process pretty_print: %v", err)
		}

		if !ok {
			newCaps = append(newCaps, cap)
			continue
		}

		switch p.Options.Direction {
		case "to":
			result := cap.Get(ppModifier).String()
			cap.SetData([]byte(result))
			newCaps = append(newCaps, cap)

		case "from":
			for _, data := range cap.Data() {
				stack = append(stack, data)

				if data == ppOpenCurlyBracket {
					count++
				}

				if data == ppCloseCurlyBracket {
					count--
				}

				if count == 0 {
					var buf bytes.Buffer
					if err := gojson.Compact(&buf, stack); err != nil {
						return nil, fmt.Errorf("process pretty_print: gojson compact: %v", err)
					}

					if json.Valid(buf.Bytes()) {
						newCap := config.NewCapsule()
						newCap.SetData(buf.Bytes())
						newCaps = append(newCaps, newCap)
					}

					stack = []byte{}
				}
			}

		default:
			return nil, fmt.Errorf("process pretty_print: direction %s: %v", p.Options.Direction, errInvalidDirection)
		}
	}

	if count != 0 {
		return nil, fmt.Errorf("process pretty_print: %d characters remain: %v", count, errPrettyPrintIncompleteJSON)
	}

	return newCaps, nil
}

/*
Apply processes encapsulated data with the PrettyPrint
processor.

Applying prettyprint formatting is handled by the
gjson PrettyPrint modifier and is applied to the root
JSON object.

This _does not_ support reversing prettyprint formatting;
this support is unnecessary for multi-line JSON objects
that are stored in a single byte array.
*/
func (p PrettyPrint) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return cap, fmt.Errorf("process pretty_print: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	switch p.Options.Direction {
	case "to":
		cap.SetData([]byte(cap.Get(ppModifier).String()))
		return cap, nil
	default:
		return cap, fmt.Errorf("process pretty_print: direction %s: %v", p.Options.Direction, errInvalidDirection)
	}
}
