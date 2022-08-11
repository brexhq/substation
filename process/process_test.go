package process

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/brexhq/substation/config"
)

var processByteTests = []struct {
	name     string
	conf     []config.Config
	test     []byte
	expected []byte
}{
	{
		"copy",
		[]config.Config{
			{
				Type: "copy",
				Settings: map[string]interface{}{
					"output_key": "foo",
				},
			},
		},
		[]byte(`bar`),
		[]byte(`{"foo":"bar"}`),
	},
	{
		"insert",
		[]config.Config{
			{
				Type: "insert",
				Settings: map[string]interface{}{
					"output_key": "foo",
					"options": map[string]interface{}{
						"value": "bar",
					},
				},
			},
		},
		[]byte(`{"hello":"world"}`),
		[]byte(`{"hello":"world","foo":"bar"}`),
	},
	{
		"gzip",
		[]config.Config{
			{
				Type: "gzip",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"direction": "from",
					},
				},
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 170, 86, 202, 72, 205, 201, 201, 87, 178, 82, 74, 207, 207, 79, 73, 170, 76, 85, 170, 5, 4, 0, 0, 255, 255, 214, 182, 196, 150, 19, 0, 0, 0},
		[]byte(`{"hello":"goodbye"}`),
	},
	{
		"base64",
		[]config.Config{
			{
				Type: "base64",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"direction": "from",
					},
				},
			},
		},
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[]byte(`{"hello":"world"}`),
	},
	{
		"split",
		[]config.Config{
			{
				Type: "split",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"separator": ".",
					},
					"input_key":  "foo",
					"output_key": "foo",
				},
			},
		},
		[]byte(`{"foo":"bar.baz"}`),
		[]byte(`{"foo":["bar","baz"]}`),
	},
	{
		"pretty_print",
		[]config.Config{
			{
				Type: "pretty_print",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"direction": "to",
					},
				},
			},
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{
  "foo": "bar"
}
`),
	},
	{
		"time",
		[]config.Config{
			{
				Type: "time",
				Settings: map[string]interface{}{
					"input_key":  "foo",
					"output_key": "foo",
					"options": map[string]interface{}{
						"input_format":  "unix",
						"output_format": "2006-01-02T15:04:05.000000Z",
					},
				},
			},
		},
		[]byte(`{"foo":1639877490}`),
		[]byte(`{"foo":"2021-12-19T01:31:30.000000Z"}`),
	},
}

func TestByterAll(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processByteTests {
		byters, err := MakeAllByters(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed, err := Byte(ctx, byters, test.test)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(processed, test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

func TestByterFactory(t *testing.T) {
	for _, test := range processByteTests {
		_, err := ByterFactory(test.conf[0])
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}

func benchmarkByterFactory(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		ByterFactory(conf)
	}
}

func BenchmarkByterFactory(b *testing.B) {
	for _, test := range processByteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkByterFactory(b, test.conf[0])
			},
		)
	}
}

func benchmarkByte(b *testing.B, conf []config.Config, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byters, _ := MakeAllByters(conf)
		Byte(ctx, byters, data)
	}
}

func BenchmarkByte(b *testing.B) {
	for _, test := range processByteTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkByte(b, test.conf, test.test)
			},
		)
	}
}

var processSliceTests = []struct {
	name     string
	conf     []config.Config
	test     [][]byte
	expected [][]byte
	err      error
}{
	{
		"aggregate",
		[]config.Config{
			{
				Type: "aggregate",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"aggregate_key": "foo",
					},
					"output_key": "aggregate.-1",
				},
			},
			{
				Type: "copy",
				Settings: map[string]interface{}{
					"input_key":  "aggregate.#",
					"output_key": "aggregate.0.count",
				},
			},
			{
				Type: "copy",
				Settings: map[string]interface{}{
					"input_key": "aggregate.0",
				},
			},
		},
		[][]byte{
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"baz"}`),
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"qux"}`),
			[]byte(`{"foo":"bar"}`),
			[]byte(`{"foo":"qux"}`),
		},
		[][]byte{
			[]byte(`{"foo":"bar","count":3}`),
			[]byte(`{"foo":"baz","count":1}`),
			[]byte(`{"foo":"qux","count":2}`),
		},
		nil,
	},
	{
		"split",
		[]config.Config{
			{
				Type: "split",
				Settings: map[string]interface{}{
					"options": map[string]interface{}{
						"separator": `\n`,
					},
				},
			},
		},
		[][]byte{
			[]byte(`foo\nbar\nbaz`),
		},
		[][]byte{
			[]byte(`foo`),
			[]byte(`bar`),
			[]byte(`baz`),
		},
		nil,
	},
}

func TestSlice(t *testing.T) {
	ctx := context.TODO()
	for _, test := range processSliceTests {
		slicers, err := MakeAllSlicers(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		res, err := Slice(ctx, slicers, test.test)
		if err != nil && errors.Is(err, test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		for i, processed := range res {
			expected := test.expected[i]
			if c := bytes.Compare(expected, processed); c != 0 {
				t.Logf("expected %s, got %s", expected, string(processed))
				t.Fail()
			}
		}
	}
}

func benchmarkSlice(b *testing.B, conf []config.Config, data [][]byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicers, _ := MakeAllSlicers(conf)
		Slice(ctx, slicers, data)
	}
}

func BenchmarkSlice(b *testing.B) {
	for _, test := range processSliceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkSlice(b, test.conf, test.test)
			},
		)
	}
}

func TestSlicerFactory(t *testing.T) {
	for _, test := range processSliceTests {
		_, err := SlicerFactory(test.conf[0])
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}
}

func benchmarkSlicerFactory(b *testing.B, conf config.Config) {
	for i := 0; i < b.N; i++ {
		SlicerFactory(conf)
	}
}

func BenchmarkSlicerFactory(b *testing.B) {
	for _, test := range processSliceTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkSlicerFactory(b, test.conf[0])
			},
		)
	}
}
