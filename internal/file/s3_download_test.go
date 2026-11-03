package file

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// fakeObject is an S3 object that is overwritten between a HeadObject and the
// download that follows it, once per entry in overwrites, as when an upload
// lands while a download is starting.
type fakeObject struct {
	versions   []string // ETag and contents of each version, in upload order
	overwrites int      // how many downloads find a newer version than their HeadObject
	idx        int
}

func (o *fakeObject) HeadObject(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	etag := o.versions[o.idx]
	return &s3.HeadObjectOutput{ETag: &etag}, nil
}

func (o *fakeObject) Download(_ context.Context, w io.WriterAt, in *s3.GetObjectInput, _ ...func(*manager.Downloader)) (int64, error) {
	if o.overwrites > 0 && o.idx < len(o.versions)-1 {
		o.overwrites--
		o.idx++
	}
	cur := o.versions[o.idx]
	if in.IfMatch == nil || *in.IfMatch != cur {
		// Write a partial part first, as the real downloader may have done
		// before a later part fails.
		if _, err := w.WriteAt([]byte("partial-old-bytes"), 0); err != nil {
			return 0, err
		}
		return 0, &smithy.GenericAPIError{Code: "PreconditionFailed"}
	}

	n, err := w.WriteAt([]byte(cur), 0)
	return int64(n), err
}

func TestDownloadS3(t *testing.T) {
	tests := []struct {
		name       string
		versions   []string
		overwrites int
		want       string
		wantErr    bool
	}{
		{
			name:     "unchanged object",
			versions: []string{"v1"},
			want:     "v1",
		},
		{
			name:       "object overwritten after HeadObject",
			versions:   []string{"v1", "v2"},
			overwrites: 1,
			want:       "v2",
		},
		{
			name:       "object keeps changing",
			versions:   []string{"v1", "v2", "v3", "v4"},
			overwrites: 3,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := os.CreateTemp(t.TempDir(), "substation")
			if err != nil {
				t.Fatal(err)
			}
			defer dst.Close()

			obj := &fakeObject{versions: tt.versions, overwrites: tt.overwrites}
			_, err = downloadS3(context.Background(), obj, obj, dst, "bucket", "key")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error when the object never settles")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(dst.Name())
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got %q, want %q (stale bytes from a failed attempt must not remain)", got, tt.want)
			}
		})
	}
}

func TestDownloadS3OtherErrorsAreNotRetried(t *testing.T) {
	dst, err := os.CreateTemp(t.TempDir(), "substation")
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()

	d := &failingDownloader{err: errors.New("access denied")}
	if _, err := downloadS3(context.Background(), &fakeObject{versions: []string{"v1"}}, d, dst, "bucket", "key"); err == nil {
		t.Fatal("expected the download error to be returned")
	}
	if d.calls != 1 {
		t.Errorf("non-precondition errors should not be retried, got %d attempts", d.calls)
	}
}

type failingDownloader struct {
	err   error
	calls int
}

func (d *failingDownloader) Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (int64, error) {
	d.calls++
	return 0, d.err
}
