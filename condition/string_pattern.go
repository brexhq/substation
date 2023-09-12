package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type stringPatternConfig struct {
	Object iconfig.Object `json:"object"`

	// Pattern is the regular expression used during inspection.
	Pattern string `json:"pattern"`
}

func (c *stringPatternConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *stringPatternConfig) Validate() error {
	if c.Pattern == "" {
		return fmt.Errorf("pattern: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newStringPattern(_ context.Context, cfg config.Config) (*stringPattern, error) {
	conf := stringPatternConfig{}
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

	insp := stringPattern{
		conf: conf,
		re:   re,
	}

	return &insp, nil
}

type stringPattern struct {
	conf stringPatternConfig

	re *regexp.Regexp
}

func (insp *stringPattern) Inspect(ctx context.Context, msg *message.Message) (output bool, err error) {
	if msg.IsControl() {
		return false, nil
	}

	if insp.conf.Object.Key == "" {
		return insp.re.Match(msg.Data()), nil
	}

	value := msg.GetValue(insp.conf.Object.Key)
	return insp.re.MatchString(value.String()), nil
}

func (c *stringPattern) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
