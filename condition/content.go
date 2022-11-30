package condition

import (
	"context"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/media"
)

/*
Content evaluates data by its content (media, MIME) type. When used in Substation pipelines, it is most effective when using processors that change the format of data (e.g., process/gzip). The inspector supports MIME types that follow this specification: https://mimesniff.spec.whatwg.org/.

The inspector has these settings:

	Type:
		media type used during inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:

	data:
		[31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0] == application/x-gzip

When loaded with a factory, the inspector uses this JSON configuration:

	{
		"type": "content",
		"settings": {
			"type": "application/x-gzip"
		}
	}
*/
type Content struct {
	Type   string `json:"type"`
	Negate bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the Content inspector.
func (c Content) Inspect(ctx context.Context, capsule config.Capsule) (output bool, err error) {
	matched := false

	media := media.Bytes(capsule.Data())
	if media == c.Type {
		matched = true
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
