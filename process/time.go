package process

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/internal/json"
)

/*
TimeOptions contain custom options settings for this processor.

InputFormat: time format of the input
InputLocation (optional): the time zone abbreviation for the input; if empty, then defaults to UTC
OutputFormat: time format of the output
OutputLocation (optional): the time zone abbreviation for the output; if empty, then defaults to UTC
*/
type TimeOptions struct {
	InputFormat    string `mapstructure:"input_format"`
	InputLocation  string `mapstructure:"output_location"`
	OutputFormat   string `mapstructure:"output_format"`
	OutputLocation string `mapstructure:"output_location"`
}

// Time implements the Byter and Channeler interfaces and converts time values between formats. More information is available in the README.
type Time struct {
	Condition condition.OperatorConfig `mapstructure:"condition"`
	Input     Input                    `mapstructure:"input"`
	Output    Output                   `mapstructure:"output"`
	Options   TimeOptions              `mapstructure:"options"`
}

// Channel processes a data channel of bytes with this processor. Conditions can be optionally applied on the channel data to enable processing.
func (p Time) Channel(ctx context.Context, ch <-chan []byte) (<-chan []byte, error) {
	var array [][]byte

	op, err := condition.OperatorFactory(p.Condition)
	if err != nil {
		return nil, err
	}

	for data := range ch {
		ok, err := op.Operate(data)
		if err != nil {
			return nil, err
		}

		if !ok {
			array = append(array, data)
			continue
		}

		processed, err := p.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
		array = append(array, processed)
	}

	output := make(chan []byte, len(array))
	for _, x := range array {
		output <- x
	}
	close(output)
	return output, nil

}

// Byte processes a byte slice with this processor
func (p Time) Byte(ctx context.Context, data []byte) ([]byte, error) {
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
		o, err := p.time(value)
		if err != nil {
			return nil, err
		}
		return json.Set(data, p.Output.Key, o)
	}

	var array []interface{}
	for _, v := range value.Array() {
		o, err := p.time(v)
		if err != nil {
			return nil, err
		}
		array = append(array, o)
	}

	return json.Set(data, p.Output.Key, array)
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
	} else if p.Options.InputFormat == "unix_nano" {
		timeNum := v.Int()
		timeDate := time.Unix(0, timeNum)
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
			return "", fmt.Errorf("err Time processor failed to parse output location: %v", err)
		}

		timeDate, err = time.ParseInLocation(p.Options.InputFormat, timeStr, loc)
		if err != nil {
			return "", fmt.Errorf("err Time processor failed to parse time as %s from %s using location %s: %v", p.Options.InputFormat, timeStr, p.Options.InputLocation, err)
		}
	} else {
		timeDate, err = time.Parse(p.Options.InputFormat, timeStr)
		if err != nil {
			return "", fmt.Errorf("err Time processor failed to parse time as %s from %s: %v", p.Options.InputFormat, timeStr, err)
		}
	}

	timeDate = timeDate.UTC()
	if p.Options.OutputLocation != "" {
		loc, err := time.LoadLocation(p.Options.OutputLocation)
		if err != nil {
			return "", fmt.Errorf("err Time processor failed to parse output location: %v", err)
		}

		timeDate = timeDate.In(loc)
	}

	ts := timeDate.Format(p.Options.OutputFormat)
	return ts, nil
}
