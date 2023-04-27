package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

// expand processes data by creating new objects from objects in arrays.
//
// This processor supports the data and object handling patterns.
type procExpand struct {
	process
}

// Create a new expand processor.
func newProcExpand(ctx context.Context, cfg config.Config) (p procExpand, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procExpand{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procExpand{}, err
	}

	return p, nil
}

// String returns the processor settings as an object.
func (p procExpand) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procExpand) Close(context.Context) error {
	return nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procExpand) Stream(ctx context.Context, in, out *config.Channel) error {
	defer out.Close()

	for capsule := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if ok, err := p.operator.Operate(ctx, capsule); err != nil {
				return fmt.Errorf("process: expand: %v", err)
			} else if !ok {
				out.Send(capsule)
				continue
			}

			// data is processed by retrieving and iterating an
			// array containing JSON objects (result) and setting
			// any remaining keys from the object (remains) into
			// each new object. if there is no Key, then the input
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
			var result, remains json.Result

			if p.Key != "" {
				result = json.Get(capsule.Data(), p.Key)
				// deleting the key from the object speeds
				// up processing large objects.
				if err := capsule.Delete(p.Key); err != nil {
					return fmt.Errorf("process: expand: %v", err)
				}

				remains = json.Get(capsule.Data(), "@this")
			} else {
				// remains is unused when there is no key
				result = json.Get(capsule.Data(), "@this")
			}

			for _, res := range result.Array() {
				// retains metadata from the original event
				newCapsule := capsule
				newCapsule.SetData([]byte{})

				// data processing
				//
				// elements from the array become new data.
				if p.Key == "" {
					newCapsule.SetData([]byte(res.String()))
					out.Send(newCapsule)
					continue
				}

				// object processing
				//
				// remaining keys from the original object are added
				// to the new object.
				for key, val := range remains.Map() {
					if err := newCapsule.Set(key, val); err != nil {
						return fmt.Errorf("process: expand: %v", err)
					}
				}

				if p.SetKey != "" {
					if err := newCapsule.Set(p.SetKey, res); err != nil {
						return fmt.Errorf("process: expand: %v", err)
					}

					out.Send(newCapsule)
					continue
				}

				// at this point there should be two objects that need to be
				// merged into a single object. the objects are merged using
				// the GJSON @join function, which joins all objects that are
				// in an array. if the array contains non-object data, then
				// it is ignored.
				//
				// [{"foo":"bar"},{"baz":"qux"}}] becomes {"foo":"bar","baz":"qux"}
				// [{"foo":"bar"},{"baz":"qux"},"quux"] becomes {"foo":"bar","baz":"qux"}
				tmp := fmt.Sprintf(`[%s,%s]`, newCapsule.Data(), res.String())
				join := json.Get([]byte(tmp), "@join")
				newCapsule.SetData([]byte(join.String()))

				out.Send(newCapsule)
			}
		}
	}

	return nil
}

// Batch processes one or more capsules with the processor.
func (p procExpand) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	newCapsules := newBatch(&capsules)
	for _, capsule := range capsules {
		if ok, err := p.operator.Operate(ctx, capsule); err != nil {
			return nil, fmt.Errorf("process: expand: %v", err)
		} else if !ok {
			newCapsules = append(newCapsules, capsule)
			continue
		}

		// data is processed by retrieving and iterating an
		// array containing JSON objects (result) and setting
		// any remaining keys from the object (remains) into
		// each new object. if there is no Key, then the input
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
		var result, remains json.Result

		if p.Key != "" {
			result = json.Get(capsule.Data(), p.Key)

			// deleting the key from the object speeds
			// up processing large objects.
			if err := capsule.Delete(p.Key); err != nil {
				return nil, fmt.Errorf("process: expand: %v", err)
			}

			remains = json.Get(capsule.Data(), "@this")
		} else {
			// remains is unused when there is no key
			result = json.Get(capsule.Data(), "@this")
		}

		for _, res := range result.Array() {
			// retains metadata from the original event
			newCapsule := capsule
			newCapsule.SetData([]byte{})

			// data processing
			//
			// elements from the array become new data.
			if p.Key == "" {
				newCapsule.SetData([]byte(res.String()))
				newCapsules = append(newCapsules, newCapsule)
				continue
			}

			// object processing
			//
			// remaining keys from the original object are added
			// to the new object.
			for key, val := range remains.Map() {
				if err := newCapsule.Set(key, val); err != nil {
					return nil, fmt.Errorf("process: expand: %v", err)
				}
			}

			if p.SetKey != "" {
				if err := newCapsule.Set(p.SetKey, res); err != nil {
					return nil, fmt.Errorf("process: expand: %v", err)
				}

				newCapsules = append(newCapsules, newCapsule)
				continue
			}

			// at this point there should be two objects that need to be
			// merged into a single object. the objects are merged using
			// the GJSON @join function, which joins all objects that are
			// in an array. if the array contains non-object data, then
			// it is ignored.
			//
			// [{"foo":"bar"},{"baz":"qux"}}] becomes {"foo":"bar","baz":"qux"}
			// [{"foo":"bar"},{"baz":"qux"},"quux"] becomes {"foo":"bar","baz":"qux"}
			tmp := fmt.Sprintf(`[%s,%s]`, newCapsule.Data(), res.String())
			join := json.Get([]byte(tmp), "@join")
			newCapsule.SetData([]byte(join.String()))

			newCapsules = append(newCapsules, newCapsule)
		}
	}

	return newCapsules, nil
}
