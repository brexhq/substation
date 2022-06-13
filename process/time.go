package process

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// TimeInvalidSettings is returned when the Time processor is configured with invalid Input and Output settings.
const TimeInvalidSettings = errors.Error("TimeInvalidSettings")

/*
TimeOptions contains custom options for the Time processor:
	InputFormat:
		time format of the input
		must be one of:
			pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
			unix: epoch
			unix_milli: epoch milliseconds
			unix_nano: epoch nanoseconds
			now: current time
	InputLocation (optional):
		the time zone abbreviation for the input
		defaults to UTC
	OutputFormat:
		time format of the output
		must be one of:
			pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
	InputLocation (optional):
		the time zone abbreviation for the output
		defaults to UTC
*/
type TimeOptions struct {
	InputFormat    string `json:"input_format"`
	InputLocation  string `json:"input_location"`
	OutputFormat   string `json:"output_format"`
	OutputLocation string `json:"output_location"`
}

/*
Time processes data by converting time values between formats. The processor supports these patterns:
	json:
		{"time":1639877490.061} >>> {"time":"2021-12-19T01:31:30.000000Z"}
	json array:
		{"time":[1639877490.061,1651705967]} >>> {"time":["2021-12-19T01:31:30.000000Z","2022-05-04T23:12:47.000000Z"]}

The processor uses this Jsonnet configuration:
	{
		type: 'time',
		settings: {
			input: {
				key: 'time',
			},
			output: {
				key: 'time',
			}
			options: {
				input_format: 'unix',
				output_format: '2006-01-02T15:04:05',
			}
		},
	}
*/
type Time struct {
	Condition condition.OperatorConfig `json:"condition"`
	Input     Input                    `json:"input"`
	Output    Output                   `json:"output"`
	Options   TimeOptions              `json:"options"`
}

// Slice processes a slice of bytes with the Time processor. Conditions are optionally applied on the bytes to enable processing.
func (p Time) Slice(ctx context.Context, s [][]byte) ([][]byte, error) {
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

// Byte processes bytes with the Time processor.
func (p Time) Byte(ctx context.Context, data []byte) ([]byte, error) {
	// "now" processing, supports json and data
	if p.Options.InputFormat == "now" {
		ts := time.Now().Format(p.Options.OutputFormat)

		if p.Output.Key != "" {
			return json.Set(data, p.Output.Key, ts)
		}

		return []byte(ts), nil
	}

	// json processing
	if p.Input.Key != "" && p.Output.Key != "" {
		if p.Options.InputFormat == "now" {
			timeDate := time.Now()
			ts := timeDate.Format(p.Options.OutputFormat)

			return json.Set(data, p.Output.Key, ts)
		}

		value := json.Get(data, p.Input.Key)

		// return input, otherwise time defaults to 1970
		if value.Type.String() == "Null" {
			return data, nil
		}

		if !value.IsArray() {
			ts, err := p.time(value)
			if err != nil {
				return nil, fmt.Errorf("byter settings %v: %v", p, err)
			}
			return json.Set(data, p.Output.Key, ts)
		}

		// json array processing
		var array []interface{}
		for _, v := range value.Array() {
			ts, err := p.time(v)
			if err != nil {
				return nil, fmt.Errorf("byter settings %v: %v", p, err)
			}
			array = append(array, ts)
		}

		return json.Set(data, p.Output.Key, array)
	}

	return nil, fmt.Errorf("byter settings %v: %v", p, TimeInvalidSettings)
}

func (p Time) time(v json.Result) (interface{}, error) {
	// epoch conversion requires special cases
	if p.Options.InputFormat == "unix" {
		timeNum := v.Int()
		timeDate := time.Unix(timeNum, 0)
		ts := timeDate.Format(p.Options.OutputFormat)
		return ts, nil
	} else if p.Options.InputFormat == "unix_milli" {
		timeNum := v.Int()
		timeDate := time.Unix(0, timeNum*1000000)
		ts := timeDate.Format(p.Options.OutputFormat)
		return ts, nil
	}

	// default time input format
	if p.Options.InputFormat == "" {
		p.Options.InputFormat = time.RFC3339
	}

	timeStr := v.String()
	var timeDate time.Time
	var err error
	if p.Options.InputLocation != "" {
		loc, err := time.LoadLocation(p.Options.InputLocation)
		if err != nil {
			return nil, fmt.Errorf("time location %s: %v", p.Options.InputLocation, err)
		}

		timeDate, err = time.ParseInLocation(p.Options.InputFormat, timeStr, loc)
		if err != nil {
			return nil, fmt.Errorf("time parse format %s location %s: %v", p.Options.InputFormat, p.Options.InputLocation, err)
		}
	} else {
		timeDate, err = time.Parse(p.Options.InputFormat, timeStr)
		if err != nil {
			return nil, fmt.Errorf("time parse format %s: %v", p.Options.InputFormat, err)
		}
	}

	timeDate = timeDate.UTC()
	if p.Options.OutputLocation != "" {
		loc, err := time.LoadLocation(p.Options.OutputLocation)
		if err != nil {
			return nil, fmt.Errorf("time location %s: %v", p.Options.OutputLocation, err)
		}

		timeDate = timeDate.In(loc)
	}

	// epoch conversion requires special cases
	if p.Options.OutputFormat == "unix" {
		return timeDate.Unix(), nil
	} else if p.Options.OutputFormat == "unix_milli" {
		return timeDate.UnixMilli(), nil
	}

	ts := timeDate.Format(p.Options.OutputFormat)
	return ts, nil
}
