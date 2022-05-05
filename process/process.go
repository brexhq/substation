package process

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/brexhq/substation/internal/errors"
)

// ByteInvalidFactoryConfig is used when an unsupported Byte is referenced in ByteFactory
const ByteInvalidFactoryConfig = errors.Error("ByteInvalidFactoryConfig")

// ChannelInvalidFactoryConfig is used when an unsupported Channel is referenced in ChannelFactory
const ChannelInvalidFactoryConfig = errors.Error("ChannelInvalidFactoryConfig")

// Config contains arbitrary JSON settings for Processors loaded via mapstructure.
type Config struct {
	Type     string
	Settings map[string]interface{}
}

// Input is the default input setting for processors that accept a single JSON key. This can be overriden by each processor.
type Input struct {
	Key string `mapstructure:"key"`
}

// Inputs is the default input setting for processors that accept multiple JSON keys. This can be overriden by each processor.
type Inputs struct {
	Keys []string `mapstructure:"keys"`
}

// Output is the default output setting for processors that produce a single JSON key. This can be overriden by each processor.
type Output struct {
	Key string `mapstructure:"key"`
}

// Byter is an interface for applying processors to bytes.
type Byter interface {
	Byte(context.Context, []byte) ([]byte, error)
}

// Byte accepts an array of Byters and applies all processors to the data.
func Byte(ctx context.Context, byters []Byter, data []byte) ([]byte, error) {
	var err error

	for _, byter := range byters {
		data, err = byter.Byte(ctx, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// Channeler is an interface for applying processors to channels of bytes.
type Channeler interface {
	Channel(context.Context, <-chan []byte) (<-chan []byte, error)
}

// Channel accepts a channel of bytes and applies all processors to data in the channel.
func Channel(ctx context.Context, channelers []Channeler, ch <-chan []byte) (<-chan []byte, error) {
	var err error

	for _, channeler := range channelers {
		ch, err = channeler.Channel(ctx, ch)
		if err != nil {
			return nil, err
		}
	}

	return ch, nil
}

// ReadOnlyChannel turns a write/read channel into a read-only channel.
func ReadOnlyChannel(ch chan []byte) <-chan []byte {
	return ch
}

// ByterFactory loads a Byter from a Config. This is the recommended function for retrieving ready-to-use Byters.
func ByterFactory(cfg Config) (Byter, error) {
	switch t := cfg.Type; t {
	case "base64":
		var p Base64
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("err retrieving %s from factory: %v", t, ByteInvalidFactoryConfig)
	}
}

// ChannelerFactory loads Channeler from a Config. This is the recommended function for retrieving ready-to-use Channelers.
func ChannelerFactory(cfg Config) (Channeler, error) {
	switch t := cfg.Type; t {
	case "base64":
		var p Base64
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p Count
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p Drop
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p Expand
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		mapstructure.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("err retrieving %s from factory: %v", t, ChannelInvalidFactoryConfig)
	}
}

// MakeAllByters accepts an array of Config and returns populated Byters. This a conveience function for loading many Byters.
func MakeAllByters(cfg []Config) ([]Byter, error) {
	var byters []Byter

	for _, c := range cfg {
		byter, err := ByterFactory(c)
		if err != nil {
			return nil, err
		}
		byters = append(byters, byter)
	}

	return byters, nil
}

// MakeAllChannelers accepts an array of Config and returns populated Channelers. This a conveience function for loading many Channelers.
func MakeAllChannelers(cfg []Config) ([]Channeler, error) {
	var channelers []Channeler

	for _, c := range cfg {
		channeler, err := ChannelerFactory(c)
		if err != nil {
			return nil, err
		}
		channelers = append(channelers, channeler)
	}

	return channelers, nil
}
