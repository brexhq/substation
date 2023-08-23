package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
	"github.com/tidwall/gjson"
)

type procTimeConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
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

type procTime struct {
	conf     procTimeConfig
	isObject bool
}

func (proc *procTime) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procTime) Close(context.Context) error {
	return nil
}

func newProcTime(_ context.Context, cfg config.Config) (*procTime, error) {
	conf := procTimeConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "" && conf.Format != "now") {
		return nil, fmt.Errorf("transform: proc_time: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Format == "" {
		return nil, fmt.Errorf("transform: proc_time: format: %v", errors.ErrMissingRequiredOption)
	}

	if conf.SetFormat == "" {
		return nil, fmt.Errorf("transform: proc_time: set_format: %v", errors.ErrMissingRequiredOption)
	}

	proc := procTime{
		conf:     conf,
		isObject: conf.SetKey != "",
	}

	return &proc, nil
}

func (proc *procTime) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Skip control messages.
	if message.IsControl() {
		return []*mess.Message{message}, nil
	}

	// "now" processing.
	if proc.conf.Format == "now" {
		ts := time.Now()

		var value interface{}
		switch proc.conf.SetFormat {
		case "unix":
			value = ts.Unix()
		case "unix_milli":
			value = ts.UnixMilli()
		default:
			value = ts.Format(proc.conf.SetFormat)
		}

		if proc.conf.SetKey != "" {
			if err := message.Set(proc.conf.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: proc_time: %v", err)
			}

			return []*mess.Message{message}, nil
		}

		var data []byte
		switch v := value.(type) {
		case int64:
			data = []byte(strconv.FormatInt(v, 10))
		case string:
			data = []byte(v)
		}

		msg, err := mess.New(
			mess.SetData(data),
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_time: %v", err)
		}

		return []*mess.Message{msg}, nil
	}

	if !proc.isObject {
		tmp, err := json.Set([]byte{}, "tmp", message.Data())
		if err != nil {
			return nil, fmt.Errorf("transform: proc_time: %v", err)
		}

		res := json.Get(tmp, "tmp")
		value, err := proc.process(res)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_time: %v", err)
		}

		var data []byte
		switch v := value.(type) {
		case int64:
			data = []byte(strconv.FormatInt(v, 10))
		case string:
			data = []byte(v)
		}

		msg, err := mess.New(
			mess.SetData(data),
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_time: %v", err)
		}

		return []*mess.Message{msg}, nil
	}

	result := message.Get(proc.conf.Key)

	// Return input, otherwise the time defaults to 1970.
	if !result.Exists() {
		return []*mess.Message{message}, nil
	}

	value, err := proc.process(result)
	if err != nil {
		return nil, fmt.Errorf("transform: proc_time: %v", err)
	}

	if err := message.Set(proc.conf.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: proc_time: %v", err)
	}

	return []*mess.Message{message}, nil
}

func (proc *procTime) process(result gjson.Result) (interface{}, error) {
	var timeDate time.Time
	switch proc.conf.Format {
	case "unix":
		secs := math.Floor(result.Float())
		nanos := math.Round((result.Float() - secs) * 1000000000)
		timeDate = time.Unix(int64(secs), int64(nanos))
	case "unix_milli":
		secs := math.Floor(result.Float())
		timeDate = time.Unix(0, int64(secs)*1000000)
	default:
		if proc.conf.Location != "" {
			loc, err := time.LoadLocation(proc.conf.Location)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_time: location %s: %v", proc.conf.Location, err)
			}

			timeDate, err = time.ParseInLocation(proc.conf.Format, result.String(), loc)
			if err != nil {
				return nil, fmt.Errorf("transform: time parse: format %s location %s: %v", proc.conf.Format, proc.conf.Location, err)
			}
		} else {
			var err error
			timeDate, err = time.Parse(proc.conf.Format, result.String())
			if err != nil {
				return nil, fmt.Errorf("transform: time parse: format %s: %v", proc.conf.Format, err)
			}
		}
	}

	timeDate = timeDate.UTC()
	if proc.conf.SetLocation != "" {
		loc, err := time.LoadLocation(proc.conf.SetLocation)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_time: location %s: %v", proc.conf.SetLocation, err)
		}

		timeDate = timeDate.In(loc)
	}

	switch proc.conf.SetFormat {
	case "unix":
		return timeDate.Unix(), nil
	case "unix_milli":
		return timeDate.UnixMilli(), nil
	default:
		return timeDate.Format(proc.conf.SetFormat), nil
	}
}
