package substation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/transform"
)

var errNoTransforms = fmt.Errorf("no transforms configured")

// Config is the core configuration for the application. Custom applications
// should embed this and add additional configuration options.
type Config struct {
	// Concurrency is the number of concurrent data transformation goroutines that
	// should be allowed to run. This is not enforced by Substation, but is used
	// by the calling application to limit the number of concurrent goroutines.
	Concurrency int `json:"concurrency"`
	// Transforms contains a list of data transformatons that are executed.
	Transforms []config.Config `json:"transforms"`
}

// Substation provides access to data transformation functions.
type Substation struct {
	cfg Config

	tforms []transform.Transformer
}

// New returns a new Substation instance.
func New(ctx context.Context, cfg Config) (*Substation, error) {
	if cfg.Transforms == nil {
		return nil, errNoTransforms
	}

	// If concurrency is not set, then concurrency is retrieved from the SUBSTATION_CONCURRENCY environment variable.
	// If the environment variable is not set, then set concurrency to -1, which means the number of concurrent
	// data transformation goroutines should be unbounded.
	if cfg.Concurrency == 0 {
		val, found := os.LookupEnv("SUBSTATION_CONCURRENCY")
		if found {
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}

			cfg.Concurrency = v
		} else {
			cfg.Concurrency = -1
		}
	}

	// Create transforms from the configuration.
	t, err := transform.NewTransformers(ctx, cfg.Transforms...)
	if err != nil {
		return nil, err
	}

	sub := Substation{
		cfg:    cfg,
		tforms: t,
	}

	return &sub, nil
}

// Closes all data transforms.
func (s *Substation) Close(ctx context.Context) error {
	for _, t := range s.tforms {
		if err := t.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Concurrency returns the configured concurrency.
func (s *Substation) Concurrency() int {
	return s.cfg.Concurrency
}

// Transforms returns the configured data transforms.
//
// These are safe to use concurrently.
func (s *Substation) Transforms() []transform.Transformer {
	return s.tforms
}

// String returns a JSON representation of the configuration.
func (s *Substation) String() string {
	b, err := json.Marshal(s.cfg)
	if err != nil {
		return fmt.Sprintf("substation: %v", err)
	}

	return string(b)
}
