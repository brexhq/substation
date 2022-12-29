package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var captureTests = []struct {
	name     string
	proc     _capture
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON find",
		_capture{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _captureOptions{
				Expression: "^([^@]*)@.*$",
				Type:       "find",
			},
		},
		[]byte(`{"foo":"bar@qux.corge"}`),
		[]byte(`{"foo":"bar"}`),
		nil,
	},
	{
		"JSON find_all",
		_capture{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: _captureOptions{
				Expression: "(.{1})",
				Type:       "find_all",
				Count:      3,
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"foo":["b","a","r"]}`),
		nil,
	},
	{
		"data",
		_capture{
			Options: _captureOptions{
				Expression: "^([^@]*)@.*$",
				Type:       "find",
			},
		},
		[]byte(`bar@qux.corge`),
		[]byte(`bar`),
		nil,
	},
	{
		"named_group",
		_capture{
			Options: _captureOptions{
				Type:       "named_group",
				Expression: "(?P<foo>[a-zA-Z]+) (?P<qux>[a-zA-Z]+)",
			},
		},
		[]byte(`bar quux`),
		[]byte(`{"foo":"bar","qux":"quux"}`),
		nil,
	},
}

func TestCapture(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range captureTests {
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkCapture(b *testing.B, applier _capture, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkCapture(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range captureTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkCapture(b, test.proc, capsule)
			},
		)
	}
}
