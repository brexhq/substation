package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setConfig points SUBSTATION_CONFIG at a temporary file containing cfg.
func setConfig(t *testing.T, cfg string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("SUBSTATION_CONFIG", path)
}

func TestHTTPHandler(t *testing.T) {
	// An object_copy with no keys returns each message unchanged.
	setConfig(t, `{"transforms":[{"type":"object_copy","settings":{}}]}`)

	t.Run("returns transformed messages as a JSON array", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"hello":"world"}`))
		w := httptest.NewRecorder()

		httpHandler(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body=%q)", w.Code, w.Body.String())
		}

		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("content-type = %q, want application/json", ct)
		}

		var out []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("response is not a JSON array: %v (body=%q)", err, w.Body.String())
		}

		if len(out) != 1 || out[0]["hello"] != "world" {
			t.Fatalf("output = %q, want the input echoed back", w.Body.String())
		}
	})

	t.Run("invalid JSON output returns a generic 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))
		w := httptest.NewRecorder()

		httpHandler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500 (body=%q)", w.Code, w.Body.String())
		}

		if got := strings.TrimSpace(w.Body.String()); got != "internal server error" {
			t.Fatalf("body = %q, want a generic error message", got)
		}
	})

	t.Run("oversized body returns 413", func(t *testing.T) {
		body := strings.NewReader(strings.Repeat("a", maxHTTPBody+1))
		req := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		httpHandler(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("status = %d, want 413", w.Code)
		}
	})
}
