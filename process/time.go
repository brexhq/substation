package process

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/brexhq/substation/condition"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
)

// time processes data by converting time values between formats.
//
// This processor supports the data and object handling patterns.
type procTime struct {
	process
	Options procTimeOptions `json:"options"`
}

type procTimeOptions struct {
	// Format is the time format of the data.
	//
	// Must be one of:
	//
	// - pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
	//
	// - unix: epoch (supports fractions of a second)
	//
	// - unix_milli: epoch milliseconds
	//
	// - now: current time
	Format string `json:"format"`
	// Location is the timezone abbreviation of the data.
	//
	// This is optional and defaults to UTC.
	Location string `json:"location"`
	// SetFormat is the time format of the processed data.
	//
	// Must be one of:
	//
	// - pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
	//
	// - unix: epoch (supports fractions of a second)
	//
	// - unix_milli: epoch milliseconds
	SetFormat string `json:"set_format"`
	// SetLocation is the timezone abbreviation of the processed data.
	//
	// This is optional and defaults to UTC.
	SetLocation string `json:"set_location"`
}

// String returns the processor settings as an object.
func (p procTime) String() string {
	return toString(p)
}

// Closes resources opened by the processor.
func (p procTime) Close(context.Context) error {
	return nil
}

// Create a new time processor.
func newProcTime(ctx context.Context, cfg config.Config) (p procTime, err error) {
	if err = config.Decode(cfg.Settings, &p); err != nil {
		return procTime{}, err
	}

	p.operator, err = condition.NewOperator(ctx, p.Condition)
	if err != nil {
		return procTime{}, err
	}

	// error early if required options are missing
	if p.Options.Format == "" || p.Options.SetFormat == "" {
		return procTime{}, fmt.Errorf("process: time: options %+v: %v", p.Options, errors.ErrMissingRequiredOption)
	}

	return p, nil
}

// Stream processes a pipeline of capsules with the processor.
func (p procTime) Stream(ctx context.Context, in, out *config.Channel) error {
	return streamApply(ctx, in, out, p)
}

// Batch processes one or more capsules with the processor.
func (p procTime) Batch(ctx context.Context, capsules ...config.Capsule) ([]config.Capsule, error) {
	return batchApply(ctx, capsules, p)
}

// Apply processes a capsule with the processor.
func (p procTime) Apply(ctx context.Context, capsule config.Capsule) (config.Capsule, error) {
	if ok, err := p.operator.Operate(ctx, capsule); err != nil {
		return capsule, fmt.Errorf("process: time: %v", err)
	} else if !ok {
		return capsule, nil
	}

	// "now" processing, supports json and data
	if p.Options.Format == "now" {
		ts := time.Now()

		var value interface{}
		switch p.Options.SetFormat {
		case "unix":
			value = ts.Unix()
		case "unix_milli":
			value = ts.UnixMilli()
		default:
			value = ts.Format(p.Options.SetFormat)
		}

		if p.SetKey != "" {
			if err := capsule.Set(p.SetKey, value); err != nil {
				return capsule, fmt.Errorf("process: time: %v", err)
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

		// return input, otherwise procTime defaults to 1970
		if result.Type.String() == "Null" {
			return capsule, nil
		}

		value, err := p.procTime(result)
		if err != nil {
			return capsule, fmt.Errorf("process: time: %v", err)
		}

		if err := capsule.Set(p.SetKey, value); err != nil {
			return capsule, fmt.Errorf("process: time: %v", err)
		}

		return capsule, nil
	}

	// data processing
	if p.Key == "" && p.SetKey == "" {
		tmp, err := json.Set([]byte{}, "tmp", capsule.Data())
		if err != nil {
			return capsule, fmt.Errorf("process: time: %v", err)
		}

		res := json.Get(tmp, "tmp")
		value, err := p.procTime(res)
		if err != nil {
			return capsule, fmt.Errorf("process: time: %v", err)
		}

		switch v := value.(type) {
		case int64:
			capsule.SetData([]byte(strconv.FormatInt(v, 10)))
		case string:
			capsule.SetData([]byte(v))
		}

		return capsule, nil
	}

	return capsule, fmt.Errorf("process: time: key %s set_key %s: %v", p.Key, p.SetKey, errInvalidDataPattern)
}

func (p procTime) procTime(result json.Result) (interface{}, error) {
	var timeDate time.Time
	switch p.Options.Format {
	case "unix":
		secs := math.Floor(result.Float())
		nanos := math.Round((result.Float() - secs) * 1000000000)
		timeDate = time.Unix(int64(secs), int64(nanos))
	case "unix_milli":
		secs := math.Floor(result.Float())
		timeDate = time.Unix(0, int64(secs)*1000000)
	default:
		if p.Options.Location != "" {
			loc, err := time.LoadLocation(p.Options.Location)
			if err != nil {
				return nil, fmt.Errorf("process: time: location %s: %v", p.Options.Location, err)
			}

			timeDate, err = time.ParseInLocation(p.Options.Format, result.String(), loc)
			if err != nil {
				return nil, fmt.Errorf("process: time parse: format %s location %s: %v", p.Options.Format, p.Options.Location, err)
			}
		} else {
			var err error
			timeDate, err = time.Parse(p.Options.Format, result.String())
			if err != nil {
				return nil, fmt.Errorf("process: time parse: format %s: %v", p.Options.Format, err)
			}
		}
	}

	timeDate = timeDate.UTC()
	if p.Options.SetLocation != "" {
		loc, err := time.LoadLocation(p.Options.SetLocation)
		if err != nil {
			return nil, fmt.Errorf("process: time: location %s: %v", p.Options.SetLocation, err)
		}

		timeDate = timeDate.In(loc)
	}

	switch p.Options.SetFormat {
	case "unix":
		return timeDate.Unix(), nil
	case "unix_milli":
		return timeDate.UnixMilli(), nil
	default:
		return timeDate.Format(p.Options.SetFormat), nil
	}
}
