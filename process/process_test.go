package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var processTests = []struct {
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
	for _, test := range processTests {
		applicators, err := MakeAll(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		cap := config.NewCapsule()
		cap.SetData(test.test)

		processed, err := Apply(ctx, cap, applicators...)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(processed.GetData(), test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

// func TestSlice(t *testing.T) {
// 	ctx := context.TODO()
// 	for _, test := range processTests {
// 		slice := make([][]byte, 1, 1)
// 		slice[0] = test.test

// 		slicers, err := MakeAllSlicers(test.conf)
// 		if err != nil {
// 			t.Log(err)
// 			t.Fail()
// 		}

// 		processed, err := Slice(ctx, slicers, slice)
// 		if err != nil {
// 			t.Log(err)
// 			t.Fail()
// 		}

// 		if c := bytes.Compare(processed[0], test.expected); c != 0 {
// 			t.Logf("expected %v, got %v", test.expected, processed)
// 			t.Fail()
// 		}
// 	}
// }

// func TestSliceFactory(t *testing.T) {
// 	ctx := context.TODO()

// 	for _, test := range processTests {
// 		slice := make([][]byte, 1, 1)
// 		slice[0] = test.test

// 		conf := test.conf[0]
// 		slicer, err := SlicerFactory(conf)
// 		if err != nil {
// 			t.Log(err)
// 			t.Fail()
// 		}

// 		processed, err := slicer.Slice(ctx, slice)
// 		if err != nil {
// 			t.Log(err)
// 			t.Fail()
// 		}

// 		if c := bytes.Compare(processed[0], test.expected); c != 0 {
// 			t.Logf("expected %v, got %v", test.expected, processed)
// 			t.Fail()
// 		}
// 	}
// }

// func BenchmarkByterFactory(b *testing.B) {
// 	ctx := context.TODO()
// 	for i := 0; i < b.N; i++ {
// 		applicator, err := ByterFactory(processTests[0].conf[0])
// 		if err != nil {
// 			b.Log(err)
// 			b.Fail()
// 		}

// 		applicator.Byte(ctx, processTests[0].test)
// 	}
// }

// var applicator, _ = ByterFactory(processTests[0].conf[0])

// func BenchmarkByte(b *testing.B) {
// 	ctx := context.TODO()
// 	for i := 0; i < b.N; i++ {
// 		applicator.Byte(ctx, processTests[0].test)
// 	}
// }

// func BenchmarkSlicerFactory(b *testing.B) {
// 	slice := make([][]byte, 1, 1)
// 	slice[0] = processTests[0].test

// 	ctx := context.TODO()
// 	for i := 0; i < b.N; i++ {
// 		slicer, err := SlicerFactory(processTests[0].conf[0])
// 		if err != nil {
// 			b.Log(err)
// 			b.Fail()
// 		}

// 		slicer.Slice(ctx, slice)
// 	}
// }

// var slicer, _ = SlicerFactory(processTests[0].conf[0])

// func BenchmarkSlice(b *testing.B) {
// 	slice := make([][]byte, 1, 1)
// 	slice[0] = processTests[0].test

// 	ctx := context.TODO()
// 	for i := 0; i < b.N; i++ {
// 		slicer.Slice(ctx, slice)
// 	}
// }
