package condition

import (
	"net/http"
	"strings"
)

/*
Content evaluates bytes by their content type. This inspector uses the standard library's net/http package to identify the content type of data (more information available here: https://pkg.go.dev/net/http#DetectContentType). When used in Substation pipelines, it is most effective when using processors that change the format of data (e.g., process/gzip). The inspector supports MIME types that follow this specification: https://mimesniff.spec.whatwg.org/.

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
	Type   string `mapstructure:"type"`
	Negate bool   `mapstructure:"negate"`
}

// Inspect evaluates data with the Content inspector.
func (c Content) Inspect(data []byte) (output bool, err error) {
	var matched bool

	content := http.DetectContentType(data)
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
