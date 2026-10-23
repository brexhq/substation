package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/log"
)

// maxHTTPBody is the largest request body, in bytes, that the handler accepts.
const maxHTTPBody = 32 << 20

// errInvalidJSON is returned when a transform produces invalid JSON and cannot be returned.
var errInvalidJSON = fmt.Errorf("transformed data is invalid JSON and cannot be returned")

type httpMetadata struct {
	Path    string              `json:"path"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
}

// respondHTTPError logs the error and returns a generic 500 response so that
// internal details are not exposed to the caller.
func respondHTTPError(w http.ResponseWriter, err error) {
	log.WithField("error", err).Error("http handler")
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// httpHandler transforms the request body and returns the transformed
// messages as a JSON array. Unlike the event handlers, invocation is
// synchronous, so transformed messages are returned to the caller.
func httpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	conf, err := getConfig(ctx)
	if err != nil {
		respondHTTPError(w, err)
		return
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		respondHTTPError(w, err)
		return
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		respondHTTPError(w, err)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxHTTPBody))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		respondHTTPError(w, err)
		return
	}

	metadata, err := json.Marshal(httpMetadata{
		Path:    r.URL.Path,
		Method:  r.Method,
		Headers: r.Header,
	})
	if err != nil {
		respondHTTPError(w, err)
		return
	}

	res, err := sub.Transform(ctx,
		message.New().SetData(body).SetMetadata(metadata),
		message.New().AsControl(),
	)
	if err != nil {
		respondHTTPError(w, err)
		return
	}

	output := []json.RawMessage{}
	for _, msg := range res {
		if msg.IsControl() {
			continue
		}

		if !json.Valid(msg.Data()) {
			respondHTTPError(w, errInvalidJSON)
			return
		}

		output = append(output, json.RawMessage(msg.Data()))
	}

	resp, err := json.Marshal(output)
	if err != nil {
		respondHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}
