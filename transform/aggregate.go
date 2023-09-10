package transform

import (
	"bytes"
	"fmt"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type aggregateArrayConfig struct {
	Object iconfig.Object `json:"object"`
	Buffer iconfig.Buffer `json:"buffer"`
}

func (c *aggregateArrayConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func aggToArray(data [][]byte) ([]byte, error) {
	msg := message.New()

	for _, d := range data {
		if err := msg.SetValue("array.-1", d); err != nil {
			return nil, err
		}
	}

	b := msg.GetValue("array")
	return b.Bytes(), nil
}

type aggregateStrConfig struct {
	Buffer iconfig.Buffer `json:"buffer"`

	// Separator is the string that separates messages.
	Separator string `json:"separator"`
}

func (c *aggregateStrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *aggregateStrConfig) Validate() error {
	if c.Separator == "" {
		return fmt.Errorf("separator: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func aggToStr(data [][]byte, separator []byte) []byte {
	return bytes.Join(data, separator)
}

func aggFromStr(data []byte, separator []byte) [][]byte {
	return bytes.Split(data, separator)
}
