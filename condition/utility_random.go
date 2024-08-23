package condition

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

type utilityRandomConfig struct{}

func (c *utilityRandomConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityRandom(_ context.Context, cfg config.Config) (*utilityRandom, error) {
	conf := utilityRandomConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	insp := utilityRandom{
		conf: conf,
		r:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return &insp, nil
}

type utilityRandom struct {
	conf utilityRandomConfig

	r *rand.Rand
}

func (insp *utilityRandom) Inspect(_ context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	return insp.r.Intn(2) == 1, nil
}

func (insp *utilityRandom) String() string {
	b, _ := json.Marshal(insp.conf)
	return string(b)
}
