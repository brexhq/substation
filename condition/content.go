package condition

import (
	"context"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/media"
)

// content evaluates data by its content (media, MIME) type.
// When used in Substation pipelines, it is most effective
// when using processors that change the format of data
// (e.g., process/gzip).
//
// This inspector supports the data handling pattern.
type content struct {
	condition
	Options contentOptions `json:"options"`
}

type contentOptions struct {
	// Type is the media type used for comparison during inspection. Media types follow this specification: https://mimesniff.spec.whatwg.org/.
	Type string `json:"type"`
}

// Inspect evaluates encapsulated data with the Content inspector.
func (c content) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
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
