package condition

import (
	"context"
	gojson "encoding/json"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
)

type inspJSONValidConfig struct {
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
}

type inspJSONValid struct {
	conf inspJSONValidConfig
}

func newInspJSONValid(_ context.Context, cfg config.Config) (*inspJSONValid, error) {
	conf := inspJSONValidConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	insp := inspJSONValid{
		conf: conf,
	}

	return &insp, nil
}

func (c *inspJSONValid) String() string {
	b, _ := gojson.Marshal(c.conf)
	return string(b)
}

func (c *inspJSONValid) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	matched := json.Valid(message.Data())

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
