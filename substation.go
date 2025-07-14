package substation

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
	"github.com/brexhq/substation/v2/transform"
)

//go:embed substation.libsonnet
var Library string

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

	factory transform.Factory
	tforms  []transform.Transformer
}

// New returns a new Substation instance.
func New(ctx context.Context, cfg Config, opts ...func(*Substation)) (*Substation, error) {
	if cfg.Transforms == nil {
		return nil, errNoTransforms
	}

	sub := &Substation{
		cfg:     cfg,
		factory: transform.New,
	}

	for _, o := range opts {
		o(sub)
	}

	// Create transforms from the configuration.
	for _, c := range cfg.Transforms {
		t, err := sub.factory(ctx, c)
		if err != nil {
			return nil, err
		}

		// Append the transform (t) to the list of transforms (sub.tforms).
		sub.tforms = append(sub.tforms, t)
	}

	return sub, nil
}

// WithTransformFactory implements a custom transform factory.
func WithTransformFactory(fac transform.Factory) func(*Substation) {
	return func(s *Substation) {
		s.factory = fac
	}
}

// Transform runs the configured data transformation functions on the
// provided messages.
//
// This is safe to use concurrently.
func (s *Substation) Transform(ctx context.Context, msg ...*message.Message) ([]*message.Message, error) {
	return transform.Apply(ctx, s.tforms, msg...)
}

// String returns a JSON representation of the configuration.
func (s *Substation) String() string {
	b, err := json.Marshal(s.cfg)
	if err != nil {
		return fmt.Sprintf("substation: %v", err)
	}

	return string(b)
}
