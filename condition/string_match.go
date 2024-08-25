package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type stringMatchConfig struct {
	Object iconfig.Object `json:"object"`

	// Pattern is the regular expression used during inspection.
	Pattern string `json:"pattern"`
}

func (c *stringMatchConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringMatchConfig) Validate() error {
	if c.Pattern == "" {
		return fmt.Errorf("pattern: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

func newStringMatch(_ context.Context, cfg config.Config) (*stringMatch, error) {
	conf := stringMatchConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	re, err := regexp.Compile(conf.Pattern)
	if err != nil {
		return nil, fmt.Errorf("condition: insp_regexp: %v", err)
	}

	insp := stringMatch{
		conf: conf,
		re:   re,
	}

	return &insp, nil
}

type stringMatch struct {
	conf stringMatchConfig

	re *regexp.Regexp
}

func (insp *stringMatch) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.SourceKey == "" {
		return insp.re.Match(msg.Data()), nil
	}

	value := msg.GetValue(insp.conf.Object.SourceKey)
	return insp.re.MatchString(value.String()), nil
}

func (c *stringMatch) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
