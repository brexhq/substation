package condition

import (
	"unicode/utf8"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type numberBitwiseConfig struct {
	// Value used for comparison during inspection.
	Value int64 `json:"value"`

	Object iconfig.Object `json:"object"`
}

func (c *numberBitwiseConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

type numberLengthConfig struct {
	// Value used for comparison during inspection.
	Value int `json:"value"`
	// Measurement controls how the length is measured. The inspector automatically
	// assigns measurement for objects when the key is an array.
	//
	// Must be one of:
	//
	// - byte: number of bytes
	//
	// - char: number of characters
	//
	// This is optional and defaults to byte.
	Measurement string `json:"measurement"`

	Object iconfig.Object `json:"object"`
}

func (c *numberLengthConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func numberLengthMeasurement(b []byte, measurement string) int {
	switch measurement {
	case "byte":
		return len(b)
	case "char", "rune": // rune is an alias for char
		return utf8.RuneCount(b)
	default:
		return len(b)
	}
}

type numberConfig struct {
	// Value used for comparison during inspection.
	Value float64 `json:"value"`

	Object iconfig.Object `json:"object"`
}

func (c *numberConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}
