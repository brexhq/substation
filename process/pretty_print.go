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

// errPrettyPrintIncompleteJSON is returned when the processor is given input
// that does not contain an equal number of open curly brackets ( { ) and close
// curly brackets ( } ), indicating that the input was an incomplete JSON object.
//
// The most common causes of this error are invalid input JSON
// (e.g., {{"foo":"bar"}) or using the processor with multi-core processing enabled.
const errPrettyPrintIncompleteJSON = errors.Error("incomplete JSON object")

// prettyPrint processes data by applying or reversing prettyprint formatting to objects.
//
// This processor has significant limitations when used to reverse prettyprint, including:
//
// - cannot support multi-core processing
//
// - invalid input will cause unpredictable results
//
// It is strongly recommended to _not_ use this processor unless absolutely necessary; a
// more reliable solution is to modify the source application emitting multi-line objects
// so that it outputs a single-line object instead.
//
// This processor supports the data handling pattern.
type _prettyPrint struct {
	process
	Options _prettyPrintOptions `json:"options"`
}

type _prettyPrintOptions struct {
	// Direction determines whether prettyprint formatting is
	// applied or reversed.
	//
	// Must be one of:
	//
	// - to: applies prettyprint formatting
	//
	// - from: reverses prettyprint formatting
	Direction string `json:"direction"`
}

// String returns the processor settings as an object.
func (p _prettyPrint) String() string {
	return toString(p)
}

// Close closes resources opened by the processor.
func (p _prettyPrint) Close(context.Context) error {
	return nil
}

// Batch processes one or more capsules with the processor.
//
// Applying prettyprint formatting is handled by the
// gjson PrettyPrint modifier and is applied to the root
// object.
//
// Reversing prettyprint formatting is handled by
// iterating incoming data per byte and pushing the
// bytes to a stack. When an equal number of open
// and close curly brackets ( { } ) are observed,
// then the stack of bytes has JSON compaction
// applied and the result is emitted as a new object.
func (p _prettyPrint) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
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

	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		ok, err := op.Operate(ctx, capsule)
		if err != nil {
			return nil, fmt.Errorf("process pretty_print: %v", err)
		}

		if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}

		switch p.Options.Direction {
		case "to":
			result := capsule.Get(ppModifier).String()
			capsule.SetData([]byte(result))
			newCapsules = append(newCapsules, capsule)

		case "from":
			for _, data := range capsule.Data() {
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
						newCapsule := config.NewCapsule()
						newCapsule.SetData(buf.Bytes())
						newCapsules = append(newCapsules, newCapsule)
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

	return newCapsules, nil
}

// Apply processes a capsule with the processor.
//
// Applying prettyprint formatting is handled by the
// gjson PrettyPrint modifier and is applied to the root
// object.
//
// This _does not_ support reversing prettyprint formatting;
// this support is unnecessary for multi-line objects that
// are stored in a single byte array.
func (p _prettyPrint) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.Direction == "" {
		return capsule, fmt.Errorf("process pretty_print: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	switch p.Options.Direction {
	case "to":
		capsule.SetData([]byte(capsule.Get(ppModifier).String()))
		return capsule, nil
	default:
		return capsule, fmt.Errorf("process pretty_print: direction %s: %v", p.Options.Direction, errInvalidDirection)
	}
}
