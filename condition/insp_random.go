package condition

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type inspRandomConfig struct{}

type inspRandom struct {
	conf inspRandomConfig
}

func newInspRandom(_ context.Context, cfg config.Config) (*inspRandom, error) {
	conf := inspRandomConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	insp := inspRandom{
		conf: conf,
	}

	return &insp, nil
}

func (c *inspRandom) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}

func (c *inspRandom) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	return rand.Intn(2) == 1, nil
}
