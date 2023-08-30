package transform

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/message"
)

type modTimeConfig struct {
	Object configObject `json:"object"`

	// Format is the time format of the data.
	//
	// Must be one of:
	//
	// - Pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
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
	// - Pattern-based layouts (https://gobyexample.com/time-formatting-parsing)
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

type modTime struct {
	conf  modTimeConfig
	isObj bool
}

func (tf *modTime) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modTime) Close(context.Context) error {
	return nil
}

func newModTime(_ context.Context, cfg config.Config) (*modTime, error) {
	conf := modTimeConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_time: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_time: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_time: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Format == "" {
		return nil, fmt.Errorf("transform: new_mod_time: format: %v", errors.ErrMissingRequiredOption)
	}

	if conf.SetFormat == "" {
		return nil, fmt.Errorf("transform: new_mod_time: set_format: %v", errors.ErrMissingRequiredOption)
	}

	tf := modTime{
		conf:  conf,
		isObj: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *modTime) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// "now" processing.
	if tf.conf.Format == "now" {
		ts := time.Now()

		var value interface{}
		switch tf.conf.SetFormat {
		case "unix":
			value = ts.Unix()
		case "unix_milli":
			value = ts.UnixMilli()
		default:
			value = ts.Format(tf.conf.SetFormat)
		}

		if tf.isObj {
			if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: mod_time: %v", err)
			}

			return []*message.Message{msg}, nil
		}

		var data []byte
		switch v := value.(type) {
		case int64:
			data = []byte(strconv.FormatInt(v, 10))
		case string:
			data = []byte(v)
		}

		outMsg := message.New().SetData(data).SetMetadata(msg.Metadata())
		return []*message.Message{outMsg}, nil
	}

	if !tf.isObj {
		tmp, err := json.Set([]byte{}, "tmp", msg.Data())
		if err != nil {
			return nil, fmt.Errorf("transform: mod_time: %v", err)
		}

		res := json.Get(tmp, "tmp")
		value, err := tf.process(res)
		if err != nil {
			return nil, fmt.Errorf("transform: mod_time: %v", err)
		}

		var data []byte
		switch v := value.(type) {
		case int64:
			data = []byte(strconv.FormatInt(v, 10))
		case string:
			data = []byte(v)
		}

		outMsg := message.New().SetData(data).SetMetadata(msg.Metadata())
		return []*message.Message{outMsg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key)

	// Return input, otherwise the time defaults to 1970.
	if !result.Exists() {
		return []*message.Message{msg}, nil
	}

	value, err := tf.process(result)
	if err != nil {
		return nil, fmt.Errorf("transform: mod_time: %v", err)
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_time: %v", err)
	}

	return []*message.Message{msg}, nil
}

func (tf *modTime) process(result json.Result) (interface{}, error) {
	var timeDate time.Time
	switch tf.conf.Format {
	case "unix":
		secs := math.Floor(result.Float())
		nanos := math.Round((result.Float() - secs) * 1000000000)
		timeDate = time.Unix(int64(secs), int64(nanos))
	case "unix_milli":
		secs := math.Floor(result.Float())
		timeDate = time.Unix(0, int64(secs)*1000000)
	default:
		if tf.conf.Location != "" {
			loc, err := time.LoadLocation(tf.conf.Location)
			if err != nil {
				return nil, fmt.Errorf("location %s: %v", tf.conf.Location, err)
			}

			timeDate, err = time.ParseInLocation(tf.conf.Format, result.String(), loc)
			if err != nil {
				return nil, fmt.Errorf("format %s location %s: %v", tf.conf.Format, tf.conf.Location, err)
			}
		} else {
			var err error
			timeDate, err = time.Parse(tf.conf.Format, result.String())
			if err != nil {
				return nil, fmt.Errorf("format %s: %v", tf.conf.Format, err)
			}
		}
	}

	timeDate = timeDate.UTC()
	if tf.conf.SetLocation != "" {
		loc, err := time.LoadLocation(tf.conf.SetLocation)
		if err != nil {
			return nil, fmt.Errorf("location %s: %v", tf.conf.SetLocation, err)
		}

		timeDate = timeDate.In(loc)
	}

	switch tf.conf.SetFormat {
	case "unix":
		return timeDate.Unix(), nil
	case "unix_milli":
		return timeDate.UnixMilli(), nil
	default:
		return timeDate.Format(tf.conf.SetFormat), nil
	}
}
