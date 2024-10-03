package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
)

func TestPost(t *testing.T) {
	tests := []struct {
		payload  interface{}
		expected error
	}{
		{
			payload:  []byte("test"),
			expected: nil,
		},
		{
			payload:  []byte("test"),
			expected: nil,
		},
	}

	serv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
	defer serv.Close()

	ctx := context.TODO()

	h := HTTP{
		retryablehttp.NewClient(),
	}

	for _, test := range tests {
		_, err := h.Post(ctx, serv.URL, test.payload)
		if !errors.Is(err, test.expected) {
			t.Errorf("expected %+v, got %+v", test.expected, err)
		}
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		expected []byte
	}{
		{
			expected: []byte("foo"),
		},
	}

	serv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("foo"))
		}))
	defer serv.Close()

	ctx := context.TODO()

	h := HTTP{
		retryablehttp.NewClient(),
	}

	for _, test := range tests {
		resp, err := h.Get(ctx, serv.URL)
		if err != nil {
			t.Fatalf("%v", err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if c := bytes.Compare(body, test.expected); c != 0 {
			t.Errorf("expected %+v, got %+v", test.expected, body)
		}
	}
}

func FuzzHTTPGet(f *testing.F) {
	// Seed the fuzzer with initial test cases
	f.Add("/xyz")
	// f.Add("https://invalid-url")

	f.Fuzz(func(t *testing.T, urlStr string) {
		// Validate the URL before making the request
		parsedURL, err := url.Parse(urlStr)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			// If the URL is invalid, we expect an error, so we can return early
			return
		}

		serv := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("foo"))
			}))
		defer serv.Close()

		ctx := context.TODO()

		h := HTTP{
			retryablehttp.NewClient(),
		}

		resp, err := h.Get(ctx, serv.URL+urlStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		// Check if the response body is as expected
		expected := []byte("foo")
		if c := bytes.Compare(body, expected); c != 0 {
			t.Errorf("expected %+v, got %+v", expected, body)
		}
	})
}
