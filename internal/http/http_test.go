package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
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
		if err != test.expected {
			t.Logf("expected %+v, got %+v", test.expected, err)
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
func TestSize(t *testing.T) {
	var tests = []struct {
		data   string
		repeat int
		key    string
	}{
		{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			1,
			"8Ex8TUWD3dWUMh6dUKaT",
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			2,
			"8Ex8TUWD3dWUMh6dUKaT",
		},
		{
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			50,
			"8Ex8TUWD3dWUMh6dUKaT",
		},
	}

	agg := Aggregate{}
	agg.New()

	for _, test := range tests {
		s := strings.Repeat(test.data, test.repeat)
		agg.Add(s)

		check := agg.Size()
		data := agg.Get()
		if check != len(data) {
			t.Logf("expected %v, got %v", len(data), check)
			t.Fail()
		}
	}
}
