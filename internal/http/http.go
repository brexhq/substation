package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/hashicorp/go-retryablehttp"
)

// errHTTPInvalidPayload is returned by Post when it receives an unexpected payload interface.
var errHTTPInvalidPayload = fmt.Errorf("invalid payload interface")

// Header contains a single HTTP header that can be passed to HTTP.Post. Multiple headers can be passed to HTTP.Post as a slice.
type Header struct {
	Key   string
	Value string
}

// HTTP wraps a retryable HTTP client.
type HTTP struct {
	Client *retryablehttp.Client
}

// Setup creates a retryable HTTP client.
func (h *HTTP) Setup() {
	h.Client = retryablehttp.NewClient()
}

// EnableXRay replaces the standard retryable HTTP client with an AWS XRay client. This method can be used when making HTTP calls on AWS infrastructure and should be enabled by looking for the environment variable "AWS_XRAY_DAEMON_ADDRESS".
func (h *HTTP) EnableXRay() {
	h.Client.HTTPClient = xray.Client(h.Client.HTTPClient)
}

// IsEnabled identifies if the HTTP client is enabled and ready to use. This method can be used for lazy loading the client.
func (h *HTTP) IsEnabled() bool {
	return h.Client != nil
}

// Get is a context-aware convenience function for making GET requests.
func (h *HTTP) Get(ctx context.Context, url string, headers ...Header) (*http.Response, error) {
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http get URL %s: %v", url, err)
	}

	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get URL %s: %v", url, err)
	}

	return resp, err
}

// Post is a context-aware convenience function for making POST requests. This method optionally supports custom headers.
func (h *HTTP) Post(ctx context.Context, url string, payload interface{}, headers ...Header) (resp *http.Response, err error) {
	var tmp []byte

	switch p := payload.(type) {
	case []byte:
		tmp = p
	case string:
		tmp = []byte(p)
	default:
		return nil, fmt.Errorf("http post URL %s: %v", url, errHTTPInvalidPayload)
	}

	req, err := retryablehttp.NewRequest("POST", url, tmp)
	if err != nil {
		return nil, fmt.Errorf("http post URL %s: %v", url, err)
	}

	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}

	resp, err = h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post URL %s: %v", url, err)
	}

	return resp, nil
}
