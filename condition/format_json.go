package condition

import (
	"context"
	"encoding/json"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/internal/config"
)

type formatJSONConfig struct{}

func (c *formatJSONConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newFormatJSON(_ context.Context, cfg config.Config) (*formatJSON, error) {
	conf := formatJSONConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	insp := formatJSON{
		conf: conf,
	}

	return &insp, nil
}

type formatJSON struct {
	conf formatJSONConfig
}

func (c *formatJSON) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	return json.Valid(msg.Data()), nil
}

func (c *formatJSON) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
