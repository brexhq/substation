package condition

import (
	"unicode/utf8"

	iconfig "github.com/brexhq/substation/internal/config"
)

type logicLengthConfig struct {
	Object iconfig.Object `json:"object"`

	// Length is the length that is used for comparison during inspection.
	Length int `json:"length"`
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
}

func (c *logicLengthConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func logicLengthMeasurement(b []byte, measurement string) int {
	switch measurement {
	case "byte":
		return len(b)
	case "char", "rune": // rune is an alias for char
		return utf8.RuneCount(b)
	default:
		return len(b)
	}
}
