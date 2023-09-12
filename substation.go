package substation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/transform"
)

var errNoTransforms = fmt.Errorf("no transforms configured")

// Config is the core configuration for the application. Custom applications
// should embed this and add additional configuration options.
type Config struct {
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

	// Create transforms from the configuration.
	var tforms []transform.Transformer
	for _, c := range cfg.Transforms {
		t, err := transform.New(ctx, c)
		if err != nil {
			return nil, err
		}

		tforms = append(tforms, t)
	}

	sub := Substation{
		cfg:    cfg,
		tforms: tforms,
	}

	return &sub, nil
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
