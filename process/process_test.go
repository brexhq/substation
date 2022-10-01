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

func TestApply(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range processTests {
		cap.SetData(test.test)

		applicators, err := MakeApplicators(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		result, err := Apply(ctx, cap, applicators...)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %v, got %v", test.expected, result)
			t.Fail()
		}
	}
}

func TestApplicatorFactory(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range processTests {
		cap.SetData(test.test)

		conf := test.conf[0]
		applicator, err := ApplicatorFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		result, err := applicator.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %v, got %v", test.expected, result)
			t.Fail()
		}
	}
}

func TestApplyBatch(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	for _, test := range processTests {
		cap.SetData(test.test)

		batch := make([]config.Capsule, 1)
		batch[0] = cap

		applicators, err := MakeBatchApplicators(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		result, err := ApplyBatch(ctx, batch, applicators...)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result[0].Data(), test.expected) {
			t.Logf("expected %v, got %v", test.expected, result)
			t.Fail()
		}
	}
}

func TestBatchApplicatorFactory(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()
	batch := make([]config.Capsule, 1)

	for _, test := range processTests {
		cap.SetData(test.test)
		batch[0] = cap

		conf := test.conf[0]
		applicator, err := BatchApplicatorFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		result, err := applicator.ApplyBatch(ctx, batch)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result[0].Data(), test.expected) {
			t.Logf("expected %v, got %v", test.expected, result)
			t.Fail()
		}
	}
}

func BenchmarkApplicatorFactory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ApplicatorFactory(processTests[0].conf[0])
	}
}

func BenchmarkBatchApplicatorFactory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BatchApplicatorFactory(processTests[0].conf[0])
	}
}
