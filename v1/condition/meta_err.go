package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"

	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type metaErrConfig struct {
	// Inspector used to inspect the message. If the inspector
	// throws an error, this inspector will return false.
	Inspector config.Config `json:"inspector"`
	// ErrorMessages are regular expressions that match error messages and determine
	// if the error should be caught.
	//
	// This is optional and defaults to an empty list (all errors are caught).
	ErrorMessages []string `json:"error_messages"`
}

func (c *metaErrConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *metaErrConfig) Validate() error {
	if c.Inspector.Type == "" {
		return fmt.Errorf("inspector: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newMetaErr(ctx context.Context, cfg config.Config) (*metaErr, error) {
	conf := metaErrConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	i, err := newInspector(ctx, conf.Inspector)
	if err != nil {
		return nil, fmt.Errorf("condition: meta_err: %v", err)
	}

	meta := metaErr{
		conf: conf,
		insp: i,
	}

	meta.errorMessages = make([]*regexp.Regexp, len(conf.ErrorMessages))
	for i, em := range conf.ErrorMessages {
		re, err := regexp.Compile(em)
		if err != nil {
			return nil, fmt.Errorf("condition: meta_err: %v", err)
		}

		meta.errorMessages[i] = re
	}

	return &meta, nil
}

type metaErr struct {
	conf metaErrConfig

	insp          inspector
	errorMessages []*regexp.Regexp
}

func (c *metaErr) Inspect(ctx context.Context, msg *message.Message) (bool, error) {
	if msg.IsControl() {
		return false, nil
	}

	match, err := c.insp.Inspect(ctx, msg)
	if err != nil {
		// Catch all errors.
		if len(c.errorMessages) == 0 {
			return false, nil
		}

		// Catch specific errors.
		for _, re := range c.errorMessages {
			if re.MatchString(err.Error()) {
				return false, nil
			}
		}

		return false, fmt.Errorf("condition: meta_err: %v", err)
	}

	return match, nil
}

func (c *metaErr) String() string {
	b, _ := json.Marshal(c.conf)
	return string(b)
}
