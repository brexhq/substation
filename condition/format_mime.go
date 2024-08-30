package condition

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/media"
)

type formatMIMEConfig struct {
	// Type is the media type used for comparison during inspection. Media types follow this specification: https://mimesniff.spec.whatwg.org/.
	Type string `json:"type"`

	Object iconfig.Object `json:"object"`
}

func (c *formatMIMEConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *formatMIMEConfig) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("type: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newFormatMIME(_ context.Context, cfg config.Config) (*formatMIME, error) {
	conf := formatMIMEConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	insp := formatMIME{
		conf: conf,
	}

	return &insp, nil
}

type formatMIME struct {
	conf formatMIMEConfig
}

func (c *formatMIME) Condition(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	media := media.Bytes(msg.Data())
	if media == c.conf.Type {
		return true, nil
	}

	return false, nil
}

func (c *formatMIME) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
