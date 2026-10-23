package transform

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Transformer = &sendGCPStorage{}

func TestSendGCPStorageValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "valid",
			settings: map[string]interface{}{"gcp": map[string]interface{}{"resource": "projects/_/buckets/bucket"}},
		},
		{
			name:     "invalid resource",
			settings: map[string]interface{}{"gcp": map[string]interface{}{"resource": "projects/project/topics/topic"}},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf := sendGCPStorageConfig{}
			if err := conf.Decode(test.settings); err != nil {
				t.Fatal(err)
			}

			err := conf.Validate()
			if (err != nil) != test.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

// TestSendGCPStorageBatch checks that every record in a batch is uploaded, using
// a fake GCS endpoint.
func TestSendGCPStorageBatch(t *testing.T) {
	var (
		mu       sync.Mutex
		uploaded []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasPrefix(r.URL.Path, "/upload/") {
			http.NotFound(w, r)
			return
		}

		// A single-request upload is multipart/related: the first part is the
		// object metadata and the second part is the object data.
		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for i := 0; ; i++ {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			b, err := io.ReadAll(part)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if i == 1 {
				mu.Lock()
				uploaded = append(uploaded, string(b))
				mu.Unlock()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"object","bucket":"bucket"}`))
	}))
	defer srv.Close()

	t.Setenv("STORAGE_EMULATOR_HOST", strings.TrimPrefix(srv.URL, "http://"))

	ctx := context.Background()
	tf, err := newSendGCPStorage(ctx, config.Config{
		Settings: map[string]interface{}{
			"gcp":       map[string]interface{}{"resource": "projects/_/buckets/bucket"},
			"batch":     map[string]interface{}{"count": 10},
			"file_path": map[string]interface{}{"suffix": ".json"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, d := range []string{`{"a":1}`, `{"a":2}`, `{"a":3}`} {
		if _, err := tf.Transform(ctx, message.New().SetData([]byte(d))); err != nil {
			t.Fatal(err)
		}
	}

	// A control message flushes the batch.
	if _, err := tf.Transform(ctx, message.New().AsControl()); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()

	want := `{"a":1}{"a":2}{"a":3}`
	if len(uploaded) != 1 || uploaded[0] != want {
		t.Errorf("uploaded %q, want [%q]", uploaded, want)
	}
}
