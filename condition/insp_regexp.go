package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

type inspRegExp struct {
	conf inspRegExpConfig

	re *regexp.Regexp
}

type inspRegExpConfig struct {
	// Key is the message key used during inspection.
	Key string `json:"key"`
	// Negate is a boolean that negates the inspection result.
	Negate bool `json:"negate"`
	// Expression is the regular expression used during inspection.
	Expression string `json:"expression"`
}

func newInspRegExp(_ context.Context, cfg config.Config) (*inspRegExp, error) {
	conf := inspRegExpConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Expression == "" {
		return nil, fmt.Errorf("condition: insp_regexp: expression: %v", errors.ErrMissingRequiredOption)
	}

	re, err := regexp.Compile(conf.Expression)
	if err != nil {
		return nil, fmt.Errorf("condition: insp_regexp: %v", err)
	}

	insp := inspRegExp{
		conf: conf,
		re:   re,
	}

	return &insp, nil
}

func (c *inspRegExp) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}

func (c *inspRegExp) Inspect(ctx context.Context, message *mess.Message) (output bool, err error) {
	if message.IsControl() {
		return false, nil
	}

	var matched bool
	if c.conf.Key == "" {
		matched = c.re.Match(message.Data())
	} else {
		res := message.Get(c.conf.Key).String()
		matched = c.re.MatchString(res)
	}

	if c.conf.Negate {
		return !matched, nil
	}

	return matched, nil
}
