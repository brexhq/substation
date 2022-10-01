package process

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

/*
Time processes data by converting time values between formats. The processor supports these patterns:
	JSON:
		{"time":1639877490.061} >>> {"time":"2021-12-19T01:31:30.061000Z"}
	data:
		1639877490.061 >>> 2021-12-19T01:31:30.061000Z

When loaded with a factory, the processor uses this JSON configuration:
	{
		"type": "time",
		"settings": {
			"options": {
				"input_format": "unix",
				"output_format": "2006-01-02T15:04:05.000000Z"
			},
			"input_key": "time",
			"output_key": "time"
		}
	}
*/
type Time struct {
	Options   TimeOptions      `json:"options"`
	Condition condition.Config `json:"condition"`
	InputKey  string           `json:"input_key"`
	OutputKey string           `json:"output_key"`
}

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

// ApplyBatch processes a slice of encapsulated data with the Time processor. Conditions are optionally applied to the data to enable processing.
func (p Time) ApplyBatch(ctx context.Context, caps []config.Capsule) ([]config.Capsule, error) {
	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, fmt.Errorf("time applybatch: %v", err)
	}

	caps, err = conditionallyApplyBatch(ctx, caps, op, p)
	if err != nil {
		return nil, fmt.Errorf("time applybatch: %v", err)
	}

	return caps, nil
}

// Apply processes encapsulated data with the Time processor.
func (p Time) Apply(ctx context.Context, cap config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.InputFormat == "" || p.Options.OutputFormat == "" {
		return cap, fmt.Errorf("time apply: options %+v: %v", p.Options, errProcessorMissingRequiredOptions)
	}

	// "now" processing, supports json and data
	if p.Options.InputFormat == "now" {
		ts := time.Now()

		var value interface{}
		switch p.Options.OutputFormat {
		case "unix":
			value = ts.Unix()
		case "unix_milli":
			value = ts.UnixMilli()
		default:
			value = ts.Format(p.Options.OutputFormat)
		}

		if p.OutputKey != "" {
			if err := cap.Set(p.OutputKey, value); err != nil {
				return cap, fmt.Errorf("time apply: %v", err)
			}

			return cap, nil
		}

		switch v := value.(type) {
		case int64:
			cap.SetData([]byte(strconv.FormatInt(v, 10)))
		case string:
			cap.SetData([]byte(v))
		}

		return cap, nil
	}

	// json processing
	if p.InputKey != "" && p.OutputKey != "" {
		result := cap.Get(p.InputKey)

		// return input, otherwise time defaults to 1970
		if result.Type.String() == "Null" {
			return cap, nil
		}

		value, err := p.time(result)
		if err != nil {
			return cap, fmt.Errorf("time apply: %v", err)
		}

		if err := cap.Set(p.OutputKey, value); err != nil {
			return cap, fmt.Errorf("time apply: %v", err)
		}

		return cap, nil
	}

	// data processing
	if p.InputKey == "" && p.OutputKey == "" {
		tmp, err := json.Set([]byte{}, "tmp", cap.Data())
		if err != nil {
			return cap, fmt.Errorf("time apply: %v", err)
		}

		res := json.Get(tmp, "tmp")
		value, err := p.time(res)
		if err != nil {
			return cap, fmt.Errorf("time apply: %v", err)
		}

		switch v := value.(type) {
		case int64:
			cap.SetData([]byte(strconv.FormatInt(v, 10)))
		case string:
			cap.SetData([]byte(v))
		}

		return cap, nil
	}

	return cap, fmt.Errorf("time apply: inputkey %s outputkey %s: %v", p.InputKey, p.OutputKey, errProcessorInvalidDataPattern)
}

func (p Time) time(result json.Result) (interface{}, error) {
	var timeDate time.Time
	switch p.Options.InputFormat {
	case "unix":
		secs := math.Floor(result.Float())
		nanos := math.Round((result.Float() - secs) * 1000000000)
		timeDate = time.Unix(int64(secs), int64(nanos))
	case "unix_milli":
		secs := math.Floor(result.Float())
		timeDate = time.Unix(0, int64(secs)*1000000)
	default:
		if p.Options.InputLocation != "" {
			loc, err := time.LoadLocation(p.Options.InputLocation)
			if err != nil {
				return nil, fmt.Errorf("time: location %s: %v", p.Options.InputLocation, err)
			}

			timeDate, err = time.ParseInLocation(p.Options.InputFormat, result.String(), loc)
			if err != nil {
				return nil, fmt.Errorf("time parse: format %s location %s: %v", p.Options.InputFormat, p.Options.InputLocation, err)
			}
		} else {
			var err error
			timeDate, err = time.Parse(p.Options.InputFormat, result.String())
			if err != nil {
				return nil, fmt.Errorf("time parse: format %s: %v", p.Options.InputFormat, err)
			}
		}
	}

	timeDate = timeDate.UTC()
	if p.Options.OutputLocation != "" {
		loc, err := time.LoadLocation(p.Options.OutputLocation)
		if err != nil {
			return nil, fmt.Errorf("time: location %s: %v", p.Options.OutputLocation, err)
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
