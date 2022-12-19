package process

import (
	"context"
	"fmt"
	gomath "math"
	"strconv"
	gotime "time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
)

/*
time processes data by converting time values between formats. The processor supports these patterns:

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
type time struct {
	process
	Options timeOptions `json:"options"`
}

/*
timeOptions contains custom options for the time processor:

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
		time zone abbreviation for the input
		defaults to UTC
	OutputLocation (optional):
		time zone abbreviation for the output
		defaults to UTC
*/
type timeOptions struct {
	InputFormat    string `json:"input_format"`
	OutputFormat   string `json:"output_format"`
	InputLocation  string `json:"input_location"`
	OutputLocation string `json:"output_location"`
}

// Close closes resources opened by the time processor.
func (p time) Close(context.Context) error {
	return nil
}

func (p time) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	capsules, err := conditionalApply(ctx, capsules, p.Condition, p)
	if err != nil {
		return nil, fmt.Errorf("process time: %v", err)
	}

	return capsules, nil
}

// Apply processes encapsulated data with the time processor.
func (p time) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	// error early if required options are missing
	if p.Options.InputFormat == "" || p.Options.OutputFormat == "" {
		return capsule, fmt.Errorf("process time: options %+v: %v", p.Options, errMissingRequiredOptions)
	}

	// "now" processing, supports json and data
	if p.Options.InputFormat == "now" {
		ts := gotime.Now()

		var value interface{}
		switch p.Options.OutputFormat {
		case "unix":
			value = ts.Unix()
		case "unix_milli":
			value = ts.UnixMilli()
		default:
			value = ts.Format(p.Options.OutputFormat)
		}

		if p.SetKey != "" {
			if err := capsule.Set(p.SetKey, value); err != nil {
				return capsule, fmt.Errorf("process time: %v", err)
			}

			return capsule, nil
		}

		switch v := value.(type) {
		case int64:
			capsule.SetData([]byte(strconv.FormatInt(v, 10)))
		case string:
			capsule.SetData([]byte(v))
		}

		return capsule, nil
	}

	// json processing
	if p.Key != "" && p.SetKey != "" {
		result := capsule.Get(p.Key)

		// return input, otherwise time defaults to 1970
		if result.Type.String() == "Null" {
			return capsule, nil
		}

		value, err := p.time(result)
		if err != nil {
			return capsule, fmt.Errorf("process time: %v", err)
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process time: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		tmp, err := json.Set([]byte{}, "tmp", capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process time: %v", err)
		}

		res := json.Get(tmp, "tmp")
		value, err := p.time(res)
		if err != nil {
			return capsule, fmt.Errorf("process time: %v", err)
		}

		switch v := value.(type) {
		case int64:
			capsule.SetData([]byte(strconv.FormatInt(v, 10)))
		case string:
			capsule.SetData([]byte(v))
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process time: inputkey %s outputkey %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}

func (p time) time(result json.Result) (interface{}, error) {
	var timeDate gotime.Time
	switch p.Options.InputFormat {
	case "unix":
		secs := gomath.Floor(result.Float())
		nanos := gomath.Round((result.Float() - secs) * 1000000000)
		timeDate = gotime.Unix(int64(secs), int64(nanos))
	case "unix_milli":
		secs := gomath.Floor(result.Float())
		timeDate = gotime.Unix(0, int64(secs)*1000000)
	default:
		if p.Options.InputLocation != "" {
			loc, err := gotime.LoadLocation(p.Options.InputLocation)
			if err != nil {
				return nil, fmt.Errorf("process time: location %s: %v", p.Options.InputLocation, err)
			}

			timeDate, err = gotime.ParseInLocation(p.Options.InputFormat, result.String(), loc)
			if err != nil {
				return nil, fmt.Errorf("process time parse: format %s location %s: %v", p.Options.InputFormat, p.Options.InputLocation, err)
			}
		} else {
			var err error
			timeDate, err = gotime.Parse(p.Options.InputFormat, result.String())
			if err != nil {
				return nil, fmt.Errorf("process time parse: format %s: %v", p.Options.InputFormat, err)
			}
		}
	}

	timeDate = timeDate.UTC()
	if p.Options.OutputLocation != "" {
		loc, err := gotime.LoadLocation(p.Options.OutputLocation)
		if err != nil {
			return nil, fmt.Errorf("process time: location %s: %v", p.Options.OutputLocation, err)
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
