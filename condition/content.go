package condition

import (
	"net/http"
	"strings"

	"github.com/brexhq/substation/config"
)

/*
Content evaluates encapsulated data by its content type. This inspector uses the standard library's net/http package to identify the content type of data (more information is available here: https://pkg.go.dev/net/http#DetectContentType). When used in Substation pipelines, it is most effective when using processors that change the format of data (e.g., process/gzip). The inspector supports MIME types that follow this specification: https://mimesniff.spec.whatwg.org/.

The inspector has these settings:
	Type:
		the MIME type used during inspection
	Negate (optional):
		if set to true, then the inspection is negated (i.e., true becomes false, false becomes true)
		defaults to false

The inspector supports these patterns:
	data:
		[31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 74, 203, 207, 7, 4, 0, 0, 255, 255, 33, 101, 115, 140, 3, 0, 0, 0] == application/x-gzip

The inspector uses this Jsonnet configuration:
	{
		type: 'content',
		// returns true if the bytes have a valid Gzip header
		settings: {
			type: 'application/x-gzip',
			negate: false,
		},
	}
*/
type Content struct {
	Type   string `json:"type"`
	Negate bool   `json:"negate"`
}

// Inspect evaluates encapsulated data with the Content inspector.
func (c Content) Inspect(cap config.Capsule) (output bool, err error) {
	var matched bool

	content := http.DetectContentType(cap.GetData())
	if strings.Compare(content, c.Type) == 0 {
		matched = true
	} else {
		matched = false
	}

	if c.Negate {
		return !matched, nil
	}

	return matched, nil
}
