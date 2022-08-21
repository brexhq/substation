package http

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
)

func TestPost(t *testing.T) {
	var tests = []struct {
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
		{
			payload:  1337,
			expected: HTTPInvalidPayload,
		},
	}

	serv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(200)
		}))
	defer serv.Close()

	ctx := context.TODO()

	h := HTTP{
		retryablehttp.NewClient(),
	}

	for _, test := range tests {
		_, err := h.Post(ctx, serv.URL, test.payload)
		if !errors.Is(err, test.expected) {
			t.Logf("expected %+v, got %+v", test.expected, err)
			t.Fail()
		}
	}
}
func TestGet(t *testing.T) {
	var tests = []struct {
		expected []byte
	}{
		{
			expected: []byte("foo"),
		},
	}

	serv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("foo"))
		}))
	defer serv.Close()

	ctx := context.TODO()

	h := HTTP{
		retryablehttp.NewClient(),
	}

	for _, test := range tests {
		resp, err := h.Get(ctx, serv.URL)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if c := bytes.Compare(body, test.expected); c != 0 {
			t.Logf("expected %+v, got %+v", test.expected, body)
		}
	}
}
