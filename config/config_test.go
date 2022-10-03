package config

import (
	"bytes"
	"encoding/json"
	"testing"
)

type Test struct {
	Foo string `json:"foo"`
	Baz int    `json:"baz,omitempty"`
}

func TestConfig(t *testing.T) {
	expected := Test{
		Foo: "bar",
		Baz: 1,
	}

	// simulates loading a JSON configuration file structured as a Config template
	config := `{"type":"test", "settings":{"foo":"bar", "baz":1}}`
	var cfg Config
	_ = json.Unmarshal([]byte(config), &cfg)

	// simulates how the interface factories are designed
	if cfg.Type == "test" {
		var instance Test
		_ = Decode(cfg.Settings, &instance)

		if instance != expected {
			t.Errorf("expected %+v, got %+v", expected, instance)
		}
	} else {
		t.Fail()
	}
}

var BenchmarkCfg = Config{
	Type: "test",
	Settings: map[string]interface{}{
		"foo": "a",
	},
}

func BenchmarkDecode(b *testing.B) {
	var instance Test

	for i := 0; i < b.N; i++ {
		_ = Decode(BenchmarkCfg.Settings, &instance)
	}
}

/*
Capsule Delete unit testing:

- data and metadata are added to a new Capsule

- JSON value is deleted using key

- JSON values are compared to expected
*/
var capsuleDeleteTests = []struct {
	name             string
	data             []byte
	metadata         interface{}
	dataExpected     []byte
	metadataExpected []byte
	key              string
	err              error
}{
	{
		"data",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar"}`),
		[]byte(`{"quux":"corge","grault":"garply"}`),
		"baz",
		nil,
	},
	{
		"metadata",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"grault":"garply"}`),
		"!metadata quux",
		nil,
	},
	// deletes all metadata
	{
		"metadata",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte{},
		"!metadata",
		nil,
	},
}

func TestCapsuleDelete(t *testing.T) {
	capsule := NewCapsule()
	for _, test := range capsuleDeleteTests {
		_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)
		_ = capsule.Delete(test.key)

		if !bytes.Equal(capsule.Data(), test.dataExpected) &&
			!bytes.Equal(capsule.Metadata(), test.metadataExpected) {
			t.Errorf("expected %s %s, got %s %s", test.dataExpected, test.metadataExpected, capsule.Data(), capsule.Metadata())
		}
	}
}

func benchmarkTestCapsuleDelete(b *testing.B, key string, capsule Capsule) {
	for i := 0; i < b.N; i++ {
		_ = capsule.Delete(key)
	}
}

func BenchmarkTestCapsuleDelete(b *testing.B) {
	capsule := NewCapsule()
	for _, test := range capsuleSetTests {
		b.Run(test.name,
			func(b *testing.B) {
				_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleDelete(b, test.key, capsule)
			},
		)
	}
}

/*
Capsule Get unit testing:

- data and metadata are added to a new Capsule

- JSON value is retrieved using key

- JSON value is compared to expected
*/
var capsuleGetTests = []struct {
	name     string
	data     []byte
	metadata interface{}
	expected string
	key      string
	err      error
}{
	{
		"data",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		"bar",
		"foo",
		nil,
	},
	{
		"metadata",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		"corge",
		"!metadata quux",
		nil,
	},
	{
		"metadata",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		"quux",
		"quux",
		"!metadata",
		nil,
	},
}

func TestCapsuleGet(t *testing.T) {
	capsule := NewCapsule()
	for _, test := range capsuleGetTests {
		_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)

		result := capsule.Get(test.key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestCapsuleGet(b *testing.B, key string, capsule Capsule) {
	for i := 0; i < b.N; i++ {
		capsule.Get(key)
	}
}

func BenchmarkTestCapsuleGet(b *testing.B) {
	capsule := NewCapsule()
	for _, test := range capsuleGetTests {
		b.Run(test.name,
			func(b *testing.B) {
				_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleGet(b, test.key, capsule)
			},
		)
	}
}

/*
Capsule Set unit testing:

- data and metadata are added to a new Capsule

- JSON value is set using key

- JSON values are compared to expected
*/
var capsuleSetTests = []struct {
	name             string
	data             []byte
	metadata         interface{}
	dataExpected     []byte
	metadataExpected []byte
	key              string
	value            interface{}
	err              error
}{
	{
		"data",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar","baz":"qux","waldo":"fred"}`),
		[]byte(`{"quux":"corge","grault":"garply"}`),
		"waldo",
		"fred",
		nil,
	},
	{
		"metadata",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"quux":"corge","grault":"garply","waldo":"fred"}`),
		"!metadata waldo",
		"fred",
		nil,
	},
}

func TestCapsuleSet(t *testing.T) {
	for _, test := range capsuleSetTests {
		capsule := NewCapsule()
		_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)

		_ = capsule.Set(test.key, test.value)
		if !bytes.Equal(capsule.Data(), test.dataExpected) &&
			!bytes.Equal(capsule.Metadata(), test.metadataExpected) {
			t.Errorf("expected %s %s, got %s %s", test.dataExpected, test.metadataExpected, capsule.Data(), capsule.Metadata())
		}
	}
}

func benchmarkTestCapsuleSet(b *testing.B, key string, val interface{}, capsule Capsule) {
	for i := 0; i < b.N; i++ {
		_ = capsule.Set(key, val)
	}
}

func BenchmarkTestCapsuleSet(b *testing.B) {
	capsule := NewCapsule()
	for _, test := range capsuleSetTests {
		b.Run(test.name,
			func(b *testing.B) {
				_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleSet(b, test.key, test.value, capsule)
			},
		)
	}
}

/*
Capsule SetRaw unit testing:

- data and metadata are added to a new Capsule

- JSON value is set using key

- JSON values are compared to expected
*/
var capsuleSetRawTests = []struct {
	name             string
	data             []byte
	metadata         interface{}
	dataExpected     []byte
	metadataExpected []byte
	key              string
	value            interface{}
	err              error
}{
	{
		"data",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar","baz":"qux","waldo":{"fred":["plugh","xyzzy","thud"]}}`),
		[]byte(`{"quux":"corge","grault":"garply"}`),
		"waldo",
		`{"fred":["plugh","xyzzy","thud"]}`,
		nil,
	},
	{
		"metadata",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"quux":"corge","grault":"garply","waldo":{"fred":["plugh","xyzzy","thud"]}}`),
		"!metadata waldo",
		`{"fred":["plugh","xyzzy","thud"]}`,
		nil,
	},
}

func TestCapsuleSetRaw(t *testing.T) {
	capsule := NewCapsule()
	for _, test := range capsuleSetRawTests {
		_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)
		_ = capsule.Set(test.key, test.value)

		if !bytes.Equal(capsule.Data(), test.dataExpected) &&
			!bytes.Equal(capsule.Metadata(), test.metadataExpected) {
			t.Errorf("expected %s %s, got %s %s", test.dataExpected, test.metadataExpected, capsule.Data(), capsule.Metadata())
		}
	}
}

func benchmarkTestCapsuleSetRaw(b *testing.B, key string, val interface{}, capsule Capsule) {
	for i := 0; i < b.N; i++ {
		_ = capsule.SetRaw(key, val)
	}
}

func BenchmarkTestCapsuleSetRaw(b *testing.B) {
	capsule := NewCapsule()
	for _, test := range capsuleSetRawTests {
		b.Run(test.name,
			func(b *testing.B) {
				_, _ = capsule.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleSetRaw(b, test.key, test.value, capsule)
			},
		)
	}
}

/*
Capsule SetData unit testing:

- data is added to a new Capsule

- data is compared to expected
*/
var capsuleSetDataTests = []struct {
	name     string
	data     []byte
	expected []byte
	err      error
}{
	{
		"data",
		[]byte(`{"foo":"bar","baz":"qux"}`),
		[]byte(`{"foo":"bar","baz":"qux"}`),
		nil,
	},
}

func TestCapsuleSetData(t *testing.T) {
	capsule := NewCapsule()
	for _, test := range capsuleSetDataTests {
		capsule.SetData(test.data)

		if !bytes.Equal(capsule.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, capsule.Metadata())
		}
	}
}

func benchmarkTestCapsuleSetData(b *testing.B, val []byte, capsule Capsule) {
	for i := 0; i < b.N; i++ {
		capsule.SetData(val)
	}
}

func BenchmarkTestCapsuleSetData(b *testing.B) {
	capsule := NewCapsule()
	for _, test := range capsuleSetDataTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTestCapsuleSetData(b, test.data, capsule)
			},
		)
	}
}

/*
Capsule SetMetadata unit testing:

- metadata is added to a new Capsule

- metadata is compared to expected
*/
var capsuleSetMetadataTests = []struct {
	name     string
	metadata interface{}
	expected []byte
	err      error
}{
	{
		"metadata",
		map[string]interface{}{
			"quux":   "corge",
			"grault": "garply",
		},
		// marshaling to JSON does not preserve order of keys
		// from the interface
		[]byte(`{"grault":"garply","quux":"corge"}`),
		nil,
	},
}

func TestCapsuleSetMetadata(t *testing.T) {
	capsule := NewCapsule()
	for _, test := range capsuleSetMetadataTests {
		_, _ = capsule.SetMetadata(test.metadata)

		if !bytes.Equal(capsule.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, capsule.Metadata())
		}
	}
}

func benchmarkTestCapsuleSetMetadata(b *testing.B, val interface{}, capsule Capsule) {
	for i := 0; i < b.N; i++ {
		_, _ = capsule.SetMetadata(val)
	}
}

func BenchmarkTestCapsuleSetMetadata(b *testing.B) {
	capsule := NewCapsule()
	for _, test := range capsuleSetMetadataTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTestCapsuleSetMetadata(b, test.metadata, capsule)
			},
		)
	}
}
