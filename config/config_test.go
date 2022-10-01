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
	json.Unmarshal([]byte(config), &cfg)

	// simulates how the interface factories are designed
	if cfg.Type == "test" {
		var instance Test
		Decode(cfg.Settings, &instance)

		if instance != expected {
			t.Logf("expected %+v, got %+v", expected, instance)
			t.Fail()
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
		Decode(BenchmarkCfg.Settings, &instance)
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
	cap := NewCapsule()
	for _, test := range capsuleDeleteTests {
		cap.SetData(test.data).SetMetadata(test.metadata)
		cap.Delete(test.key)

		if !bytes.Equal(cap.Data(), test.dataExpected) &&
			!bytes.Equal(cap.Metadata(), test.metadataExpected) {
			t.Logf("expected %s %s, got %s %s", test.dataExpected, test.metadataExpected, cap.Data(), cap.Metadata())
			t.Fail()
		}
	}
}

func benchmarkTestCapsuleDelete(b *testing.B, key string, cap Capsule) {
	for i := 0; i < b.N; i++ {
		cap.Delete(key)
	}
}

func BenchmarkTestCapsuleDelete(b *testing.B) {
	cap := NewCapsule()
	for _, test := range capsuleSetTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleDelete(b, test.key, cap)
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
	cap := NewCapsule()
	for _, test := range capsuleGetTests {
		cap.SetData(test.data).SetMetadata(test.metadata)

		result := cap.Get(test.key).String()
		if result != test.expected {
			t.Logf("expected %s, got %s", test.expected, result)
			t.Fail()
		}
	}
}

func benchmarkTestCapsuleGet(b *testing.B, key string, cap Capsule) {
	for i := 0; i < b.N; i++ {
		cap.Get(key)
	}
}

func BenchmarkTestCapsuleGet(b *testing.B) {
	cap := NewCapsule()
	for _, test := range capsuleGetTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleGet(b, test.key, cap)
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
		cap := NewCapsule()
		cap.SetData(test.data).SetMetadata(test.metadata)

		cap.Set(test.key, test.value)
		if !bytes.Equal(cap.Data(), test.dataExpected) &&
			!bytes.Equal(cap.Metadata(), test.metadataExpected) {
			t.Logf("expected %s %s, got %s %s", test.dataExpected, test.metadataExpected, cap.Data(), cap.Metadata())
			t.Fail()
		}
	}
}

func benchmarkTestCapsuleSet(b *testing.B, key string, val interface{}, cap Capsule) {
	for i := 0; i < b.N; i++ {
		cap.Set(key, val)
	}
}

func BenchmarkTestCapsuleSet(b *testing.B) {
	cap := NewCapsule()
	for _, test := range capsuleSetTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleSet(b, test.key, test.value, cap)
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
	cap := NewCapsule()
	for _, test := range capsuleSetRawTests {
		cap.SetData(test.data).SetMetadata(test.metadata)
		cap.Set(test.key, test.value)

		if !bytes.Equal(cap.Data(), test.dataExpected) &&
			!bytes.Equal(cap.Metadata(), test.metadataExpected) {
			t.Logf("expected %s %s, got %s %s", test.dataExpected, test.metadataExpected, cap.Data(), cap.Metadata())
			t.Fail()
		}
	}
}

func benchmarkTestCapsuleSetRaw(b *testing.B, key string, val interface{}, cap Capsule) {
	for i := 0; i < b.N; i++ {
		cap.SetRaw(key, val)
	}
}

func BenchmarkTestCapsuleSetRaw(b *testing.B) {
	cap := NewCapsule()
	for _, test := range capsuleSetRawTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.data).SetMetadata(test.metadata)
				benchmarkTestCapsuleSetRaw(b, test.key, test.value, cap)
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
	cap := NewCapsule()
	for _, test := range capsuleSetDataTests {
		cap.SetData(test.data)

		if !bytes.Equal(cap.Data(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, cap.Metadata())
			t.Fail()
		}
	}
}

func benchmarkTestCapsuleSetData(b *testing.B, val []byte, cap Capsule) {
	for i := 0; i < b.N; i++ {
		cap.SetData(val)
	}
}

func BenchmarkTestCapsuleSetData(b *testing.B) {
	cap := NewCapsule()
	for _, test := range capsuleSetDataTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkTestCapsuleSetData(b, test.data, cap)
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
	cap := NewCapsule()
	for _, test := range capsuleSetMetadataTests {
		cap.SetMetadata(test.metadata)

		if !bytes.Equal(cap.Metadata(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, cap.Metadata())
			t.Fail()
		}
	}
}

func benchmarkTestCapsuleSetMetadata(b *testing.B, val interface{}, cap Capsule) {
	for i := 0; i < b.N; i++ {
		cap.SetMetadata(val)
	}
}

func BenchmarkTestCapsuleSetMetadata(b *testing.B) {
	cap := NewCapsule()
	for _, test := range capsuleSetMetadataTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkTestCapsuleSetMetadata(b, test.metadata, cap)
			},
		)
	}
}
