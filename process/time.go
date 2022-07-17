package process

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
TimeOptions contains custom options for the Time processor:
	InputFormat:
		time format of the input
		must be one of:
			pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
			unix: epoch (supports fractions of a second)
			unix_milli: epoch milliseconds
			now: current time
	OutputFormat:
		time format of the output
		must be one of:
			pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
			unix: epoch
			unix_milli: epoch milliseconds
	InputLocation (optional):
		the time zone abbreviation for the input
		defaults to UTC
	OutputLocation (optional):
		the time zone abbreviation for the output
		defaults to UTC
*/
type TimeOptions struct {
	InputFormat    string `json:"input_format"`
	OutputFormat   string `json:"output_format"`
	InputLocation  string `json:"input_location"`
	OutputLocation string `json:"output_location"`
}

/*
Time processes data by converting time values between formats. The processor supports these patterns:
	json:
		{"time":1639877490.061} >>> {"time":"2021-12-19T01:31:30.061000Z"}
	data:
		1639877490.061 >>> 2021-12-19T01:31:30.061000Z

The processor uses this Jsonnet configuration:
	{
		type: 'time',
		settings: {
			input_key: 'time',
			output_key: 'time',
			options: {
				input_format: 'unix',
				output_format: '2006-01-02T15:04:05.000000Z',
			}
		},
	}
*/
type Time struct {
	Options   TimeOptions              `json:"options"`
	Condition condition.OperatorConfig `json:"condition"`
	InputKey  string                   `json:"input_key"`
	OutputKey string                   `json:"output_key"`
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
	// error early if required options are missing
	if p.Options.InputFormat == "" || p.Options.OutputFormat == "" {
		return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
	}

	// "now" processing, supports json and data
	if p.Options.InputFormat == "now" {
		ts := time.Now()
		var output interface{}

		switch p.Options.OutputFormat {
		case "unix":
			output = ts.Unix()
		case "unix_milli":
			output = ts.UnixMilli()
		default:
			output = ts.Format(p.Options.OutputFormat)
		}

		if p.OutputKey != "" {
			return json.Set(data, p.OutputKey, output)
		}

		switch v := output.(type) {
		case int64:
			return []byte(strconv.FormatInt(v, 10)), nil
		case string:
			return []byte(v), nil
		}
	}

	// json processing
	if p.InputKey != "" && p.OutputKey != "" {
		value := json.Get(data, p.InputKey)

		// return input, otherwise time defaults to 1970
		if value.Type.String() == "Null" {
			return data, nil
		}

		ts, err := p.time(value)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %v", p, err)
		}
		return json.Set(data, p.OutputKey, ts)
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		tmp, err := json.Set([]byte{}, "_tmp", data)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %v", p, err)
		}

		value := json.Get(tmp, "_tmp")
		ts, err := p.time(value)
		if err != nil {
			return nil, fmt.Errorf("byter settings %+v: %v", p, err)
		}

		switch v := ts.(type) {
		case int64:
			return []byte(strconv.FormatInt(v, 10)), nil
		case string:
			return []byte(v), nil
		}
	}

	return nil, fmt.Errorf("byter settings %+v: %v", p, ProcessorInvalidSettings)
}

func (p Time) time(v json.Result) (interface{}, error) {
	var timeDate time.Time
	switch p.Options.InputFormat {
	case "unix":
		secs := math.Floor(v.Float())
		nanos := math.Round((v.Float() - secs) * 1000000000)
		timeDate = time.Unix(int64(secs), int64(nanos))
	case "unix_milli":
		secs := math.Floor(v.Float())
		timeDate = time.Unix(0, int64(secs)*1000000)
	default:
		if p.Options.InputLocation != "" {
			loc, err := time.LoadLocation(p.Options.InputLocation)
			if err != nil {
				return nil, fmt.Errorf("time location %s: %v", p.Options.InputLocation, err)
			}

			timeDate, err = time.ParseInLocation(p.Options.InputFormat, v.String(), loc)
			if err != nil {
				return nil, fmt.Errorf("time parse format %s location %s: %v", p.Options.InputFormat, p.Options.InputLocation, err)
			}
		} else {
			var err error
			timeDate, err = time.Parse(p.Options.InputFormat, v.String())
			if err != nil {
				return nil, fmt.Errorf("time parse format %s: %v", p.Options.InputFormat, err)
			}
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

	switch p.Options.OutputFormat {
	case "unix":
		return timeDate.Unix(), nil
	case "unix_milli":
		return timeDate.UnixMilli(), nil
	default:
		return timeDate.Format(p.Options.OutputFormat), nil
	}
}
