package json

import (
	"bytes"
	"testing"
)

var getTests = []struct {
	name     string
	key      string
	test     []byte
	expected interface{}
}{
	{
		name:     "string",
		key:      "foo",
		test:     []byte(`{"foo":"string"}`),
		expected: "string",
	},
	{
		name:     "int",
		key:      "foo",
		test:     []byte(`{"foo":123}`),
		expected: 123.0,
	},
	{
		name:     "float",
		key:      "foo",
		test:     []byte(`{"foo":123.456}`),
		expected: 123.456,
	},
	{
		name:     "true",
		key:      "foo",
		test:     []byte(`{"foo":true}`),
		expected: true,
	},
	{
		name:     "false",
		key:      "foo",
		test:     []byte(`{"foo":false}`),
		expected: false,
	},
	{
		name: "multi-line",
		key:  "foo",
		test: []byte(`{
"foo":
"string"
}
`),
		expected: "string",
	},
}

func TestGetJson(t *testing.T) {
	for _, test := range getTests {
		result := Get(test.test, test.key)

		if result.Value() != test.expected {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
	}
}

func benchmarkGetJSON(b *testing.B, test []byte, key string) {
	for i := 0; i < b.N; i++ {
		Get(test, key)
	}
}

func BenchmarkGetJSON(b *testing.B) {
	for _, test := range getTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkGetJSON(b, test.test, test.key)
			},
		)
	}
}

var setTests = []struct {
	name     string
	key      string
	value    interface{}
	test     []byte
	expected []byte
}{
	{
		name:     "string",
		key:      "baz",
		value:    "qux",
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":"qux"}`),
	},
	{
		name:     "int",
		key:      "baz",
		value:    123,
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":123}`),
	},
	{
		name:     "float",
		key:      "baz",
		value:    123.456,
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":123.456}`),
	},
	{
		name:     "true",
		key:      "baz",
		value:    true,
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":true}`),
	},
	{
		name:     "JSON string",
		key:      "baz",
		value:    `{"qux":"quux"}`,
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
	},
	{
		name:     "JSON bytes",
		key:      "baz",
		value:    []byte(`{"qux":"quux"}`),
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
	},
	{
		name:     "bytes",
		key:      "baz",
		value:    []byte{120, 156, 5, 192, 33, 13, 0, 0, 0, 128, 176, 182, 216, 247, 119, 44, 6, 2, 130, 1, 69},
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":"eJwFwCENAAAAgLC22Pd3LAYCggFF"}`),
	},
}

func TestSetJson(t *testing.T) {
	for _, test := range setTests {
		result, err := Set(test.test, test.key, test.value)
		if err != nil {
			t.Errorf("got error %v", err)
			return
		}

		if c := bytes.Compare(result, test.expected); c != 0 {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkSetJSON(b *testing.B, test []byte, key string, value interface{}) {
	for i := 0; i < b.N; i++ {
		_, _ = Set(test, key, value)
	}
}

func BenchmarkSetJSON(b *testing.B) {
	for _, test := range setTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkSetJSON(b, test.test, test.key, test.value)
			},
		)
	}
}

var setRawTests = []struct {
	name     string
	key      string
	value    string
	test     []byte
	expected []byte
}{
	{
		name:     "JSON",
		key:      "baz",
		value:    `{"qux":"quux"}`,
		test:     []byte(`{"foo":"bar"}`),
		expected: []byte(`{"foo":"bar","baz":{"qux":"quux"}}`),
	},
}

func TestSetRawJson(t *testing.T) {
	for _, test := range setRawTests {
		result, err := SetRaw(test.test, test.key, test.value)
		if err != nil {
			t.Errorf("got error %v", err)
			return
		}

		if c := bytes.Compare(result, test.expected); c != 0 {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkSetRawJSON(b *testing.B, test []byte, key string, value interface{}) {
	for i := 0; i < b.N; i++ {
		_, _ = Set(test, key, value)
	}
}

func BenchmarkSetRawJSON(b *testing.B) {
	for _, test := range setRawTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkSetRawJSON(b, test.test, test.key, test.value)
			},
		)
	}
}

var validTests = []struct {
	name     string
	test     interface{}
	expected bool
}{
	{
		name:     "true string",
		test:     `{"foo":"bar"}`,
		expected: true,
	},
	{
		name:     "true bytes",
		test:     []byte(`{"foo":"bar"}`),
		expected: true,
	},
	{
		name:     "false string",
		test:     `{foo:"bar"}`,
		expected: false,
	},
	{
		name:     "false bytes",
		test:     []byte(`{foo:"bar"}`),
		expected: false,
	},
}

func TestValidJson(t *testing.T) {
	for _, test := range validTests {
		result := Valid(test.test)

		if result != test.expected {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
	}
}

func benchmarkValidJSON(b *testing.B, test interface{}) {
	for i := 0; i < b.N; i++ {
		Valid(test)
	}
}

func BenchmarkValidJSON(b *testing.B) {
	for _, test := range validTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkValidJSON(b, test.test)
			},
		)
	}
}
