package process

import (
	"bytes"
	"context"
	"testing"
)

func TestCapture(t *testing.T) {

	var tests = []struct {
		proc     Capture
		test     []byte
		expected []byte
	}{
		{
			Capture{
				Input: Input{
					Key: "capture",
				},
				Options: CaptureOptions{
					Expression: "^([^@]*)@.*$",
					Count:      1,
				},
				Output: Output{
					Key: "capture",
				},
			},
			[]byte(`{"capture":"jd@foo.com"}`),
			[]byte(`{"capture":"jd"}`),
		},
		{
			Capture{
				Input: Input{
					Key: "capture",
				},
				Options: CaptureOptions{
					Expression: "([^,]+)",
					Count:      -1,
				},
				Output: Output{
					Key: "capture",
				},
			},
			[]byte(`{"capture":"this,is,a,csv"}`),
			[]byte(`{"capture":["this","is","a","csv"]}`),
		},
		{
			Capture{
				Input: Input{
					Key: "capture",
				},
				Options: CaptureOptions{
					Expression: "(.{1})",
					Count:      5,
				},
				Output: Output{
					Key: "capture",
				},
			},
			[]byte(`{"capture":"hello"}`),
			[]byte(`{"capture":["h","e","l","l","o"]}`),
		},
		// array support
		{
			Capture{
				Input: Input{
					Key: "capture",
				},
				Options: CaptureOptions{
					Expression: "^([^@]*)@.*$",
					Count:      1,
				},
				Output: Output{
					Key: "capture",
				},
			},
			[]byte(`{"capture":["jd@foo.com","fd@foo.com"]}`),
			[]byte(`{"capture":["jd","fd"]}`),
		},
		{
			Capture{
				Input: Input{
					Key: "capture",
				},
				Options: CaptureOptions{
					Expression: "(.{1})",
					Count:      5,
				},
				Output: Output{
					Key: "capture",
				},
			},
			[]byte(`{"capture":["hello","goodbye"]}`),
			[]byte(`{"capture":[["h","e","l","l","o"],["g","o","o","d","b"]]}`),
		},
	}

	for _, test := range tests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}
