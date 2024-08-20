package transform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	gohttp "net/http"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/errors"
)

// enrichHTTPInterp is used for interpolating data into URLs.
const enrichHTTPInterp = `${DATA}`

type enrichDNSConfig struct {
	ID      string          `json:"id"`
	Object  iconfig.Object  `json:"object"`
	Request iconfig.Request `json:"request"`
}

func (c *enrichDNSConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *enrichDNSConfig) Validate() error {
	if c.Object.SourceKey == "" && c.Object.TargetKey != "" {
		return fmt.Errorf("object_source_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Object.SourceKey != "" && c.Object.TargetKey == "" {
		return fmt.Errorf("object_target_key: %v", errors.ErrMissingRequiredOption)
	}

	if c.Request.Timeout == "" {
		c.Request.Timeout = "1s"
	}

	return nil
}

func enrichHTTPParseResponse(resp *gohttp.Response) ([]byte, error) {
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	dst := &bytes.Buffer{}
	if json.Valid(buf) {
		// Compact converts a multi-line object into a single-line object.
		if err := json.Compact(dst, buf); err != nil {
			return nil, err
		}
	} else {
		dst = bytes.NewBuffer(buf)
	}

	return dst.Bytes(), nil
}
