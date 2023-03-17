package condition

import (
	"context"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/internal/media"
)

// content evaluates data by its content (media, MIME) type.
// When used in Substation pipelines, it is most effective
// when using processors that change the format of data
// (e.g., process/gzip).
//
// This inspector supports the data handling pattern.
type inspContent struct {
	condition
	Options inspContentOptions `json:"options"`
}

type inspContentOptions struct {
	// Type is the media type used for comparison during inspection. Media types follow this specification: https://mimesniff.spec.whatwg.org/.
	Type string `json:"type"`
}

// Creates a new content inspector.
func newInspContent(cfg config.Config) (c inspContent, err error) {
	err = config.Decode(cfg.Settings, &c)
	if err != nil {
		return inspContent{}, err
	}

	if c.Options.Type == "" {
		return inspContent{}, fmt.Errorf("condition: content: type missing: %w", errors.ErrMissingRequiredOptions)
	}

	return c, nil
}

func (c inspContent) String() string {
	return toString(c)
}

// Inspect evaluates encapsulated data with the content inspector.
func (c inspContent) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	matched := false

	media := media.Bytes(capsule.Data())
	if media == c.Options.Type {
		matched = true
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
