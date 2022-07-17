package process

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

// ProcessorInvalidSettings is returned when a processor is configured with invalid settings. Common causes include improper input and output settings (e.g., missing keys) and missing required options.
const ProcessorInvalidSettings = errors.Error("ProcessorInvalidSettings")

// ProcessorInvalidDirection is returned when an invalid direction option is configured on a processor.
const ProcessorInvalidDirection = errors.Error("ProcessorInvalidDirection")

// ByteInvalidFactoryConfig is used when an unsupported Byte is referenced in ByteFactory.
const ByteInvalidFactoryConfig = errors.Error("ByteInvalidFactoryConfig")

// SliceInvalidFactoryConfig is used when an unsupported Slice is referenced in SliceFactory.
const SliceInvalidFactoryConfig = errors.Error("SliceInvalidFactoryConfig")

// Slicer is an interface for applying processors to slices of bytes.
type Slicer interface {
	Slice(context.Context, [][]byte) ([][]byte, error)
}

// Slice accepts an array of Slicers and applies all processors to the data.
func Slice(ctx context.Context, slicers []Slicer, slice [][]byte) ([][]byte, error) {
	var err error

	for _, slicer := range slicers {
		slice, err = slicer.Slice(ctx, slice)
		if err != nil {
			return nil, err
		}
	}

	return slice, nil
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

// ByterFactory loads a Byter from a Config. This is the recommended function for retrieving ready-to-use Byters.
func ByterFactory(cfg config.Config) (Byter, error) {
	switch t := cfg.Type; t {
	case "base64":
		var p Base64
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p ForEach
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p Pipeline
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %v: %v", cfg.Settings, ByteInvalidFactoryConfig)
	}
}

// SlicerFactory loads a Slicer from a Config. This is the recommended function for retrieving ready-to-use Slicers.
func SlicerFactory(cfg config.Config) (Slicer, error) {
	switch t := cfg.Type; t {
	case "base64":
		var p Base64
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "capture":
		var p Capture
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "case":
		var p Case
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "concat":
		var p Concat
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "convert":
		var p Convert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "copy":
		var p Copy
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "count":
		var p Count
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "delete":
		var p Delete
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "domain":
		var p Domain
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "drop":
		var p Drop
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "dynamodb":
		var p DynamoDB
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "expand":
		var p Expand
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "flatten":
		var p Flatten
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "for_each":
		var p ForEach
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "group":
		var p Group
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "gzip":
		var p Gzip
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "hash":
		var p Hash
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "insert":
		var p Insert
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "lambda":
		var p Lambda
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "math":
		var p Math
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "pipeline":
		var p Pipeline
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "replace":
		var p Replace
		config.Decode(cfg.Settings, &p)
		return p, nil
	case "time":
		var p Time
		config.Decode(cfg.Settings, &p)
		return p, nil
	default:
		return nil, fmt.Errorf("process settings %v: %v", cfg.Settings, SliceInvalidFactoryConfig)
	}
}

// MakeAllByters accepts an array of Config and returns populated Byters. This a conveience function for loading many Byters.
func MakeAllByters(cfg []config.Config) ([]Byter, error) {
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

// MakeAllSlicers accepts an array of Config and returns populated Slicers. This a conveience function for loading many Slicers.
func MakeAllSlicers(cfg []config.Config) ([]Slicer, error) {
	var slicers []Slicer

	for _, c := range cfg {
		slicer, err := SlicerFactory(c)
		if err != nil {
			return nil, err
		}
		slicers = append(slicers, slicer)
	}

	return slicers, nil
}

// NewSlice returns a byte slice with a minimum capacity of 10.
func NewSlice(s *[][]byte) [][]byte {
	if len(*s) > 10 {
		return make([][]byte, 0, len(*s))
	}
	return make([][]byte, 0, 10)
}
