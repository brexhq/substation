package condition

import (
	"context"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/media"
	mess "github.com/brexhq/substation/message"
)

type inspContentConf struct {
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
	// Type is the media type used for comparison during inspection. Media types follow this specification: https://mimesniff.spec.whatwg.org/.
	Type string `json:"type"`
}

type inspContent struct {
	conf inspContentConf
}

func newInspContent(_ context.Context, cfg config.Config) (*inspContent, error) {
	conf := inspContentConf{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Type == "" {
		return nil, fmt.Errorf("condition: insp_content: type: %v", errors.ErrMissingRequiredOption)
	}

	insp := inspContent{
		conf: conf,
	}

	return &insp, nil
}

func (c *inspContent) String() string {
	b, _ := gojson.Marshal(c.conf)
	return string(b)
}

func (c *inspContent) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	matched := false

	media := media.Bytes(message.Data())
	if media == c.conf.Type {
		matched = true
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
