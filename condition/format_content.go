package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/media"
	"github.com/brexhq/substation/message"
)

type formatContentConfig struct {
	Object iconfig.Object `json:"object"`

	// Type is the media type used for comparison during inspection. Media types follow this specification: https://mimesniff.spec.whatwg.org/.
	Type string `json:"type"`
}

func (c *formatContentConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *formatContentConfig) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("type: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newFormatContent(_ context.Context, cfg config.Config) (*formatContent, error) {
	conf := formatContentConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	insp := formatContent{
		conf: conf,
	}

	return &insp, nil
}

type formatContent struct {
	conf formatContentConfig
}

func (c *formatContent) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	media := media.Bytes(msg.Data())
	if media == c.conf.Type {
		return true, nil
	}

	return false, nil
}

func (c *formatContent) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
