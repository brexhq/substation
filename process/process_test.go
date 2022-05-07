package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/internal/config"
)

var processTests = []struct {
	conf     []config.Config
	test     []byte
	expected []byte
}{
	{
		[]config.Config{
			{
				Type: "insert",
				Settings: map[string]interface{}{
					"condition": struct {
						Operator string
					}{
						Operator: "all",
					},
					"options": struct {
						Value interface{}
					}{
						Value: "bar",
					},
					"output": struct {
						Key string
					}{
						Key: "foo",
					},
				},
			},
		},
		[]byte(`{"hello":"world"}`),
		[]byte(`{"hello":"world","foo":"bar"}`),
	},
	{
		[]config.Config{
			{
				Type: "gzip",
				Settings: map[string]interface{}{
					"condition": struct {
						Operator string
					}{
						Operator: "all",
					},
					"options": struct {
						Direction string
					}{
						Direction: "from",
					},
				},
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 170, 86, 202, 72, 205, 201, 201, 87, 178, 82, 74, 207, 207, 79, 73, 170, 76, 85, 170, 5, 4, 0, 0, 255, 255, 214, 182, 196, 150, 19, 0, 0, 0},
		[]byte(`{"hello":"goodbye"}`),
	},
	{
		[]config.Config{
			{
				Type: "base64",
				Settings: map[string]interface{}{
					"condition": struct {
						Operator string
					}{
						Operator: "all",
					},
					"options": struct {
						Direction string
						Alphabet  string
					}{
						Direction: "from",
						Alphabet:  "std",
					},
				},
			},
		},
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[]byte(`{"hello":"world"}`),
	},
}

func TestByterAll(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processTests {
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
	ctx := context.TODO()

	for _, test := range processTests {
		conf := test.conf[0]
		byter, err := ByterFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed, err := byter.Byte(ctx, test.test)
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

func TestSlice(t *testing.T) {
	ctx := context.TODO()
	for _, test := range processTests {
		slice := make([][]byte, 1, 1)
		slice[0] = test.test

		slicers, err := MakeAllSlicers(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed, err := Slice(ctx, slicers, slice)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(processed[0], test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

func TestSliceFactory(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processTests {
		slice := make([][]byte, 1, 1)
		slice[0] = test.test

		conf := test.conf[0]
		slicer, err := SlicerFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed, err := slicer.Slice(ctx, slice)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(processed[0], test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

func BenchmarkByterFactory(b *testing.B) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter, err := ByterFactory(processTests[0].conf[0])
		if err != nil {
			b.Log(err)
			b.Fail()
		}

		byter.Byte(ctx, processTests[0].test)
	}
}

var byter, _ = ByterFactory(processTests[0].conf[0])

func BenchmarkByte(b *testing.B) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, processTests[0].test)
	}
}

func BenchmarkSlicerFactory(b *testing.B) {
	slice := make([][]byte, 1, 1)
	slice[0] = processTests[0].test

	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer, err := SlicerFactory(processTests[0].conf[0])
		if err != nil {
			b.Log(err)
			b.Fail()
		}

		slicer.Slice(ctx, slice)
	}
}

var slicer, _ = SlicerFactory(processTests[0].conf[0])

func BenchmarkSlice(b *testing.B) {
	slice := make([][]byte, 1, 1)
	slice[0] = processTests[0].test

	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		slicer.Slice(ctx, slice)
	}
}
