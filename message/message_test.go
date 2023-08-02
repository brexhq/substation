package message

import (
	"bytes"
	"strings"
	"testing"
)

var messageDeleteTests = []struct {
	name     string
	data     []byte
	expected []byte
	key      string
}{
	{
		"a",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"c":"d"}`),
		"a",
	},
	{
		"@this",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte{},
		"@this",
	},
}

func TestMessageDeleteData(t *testing.T) {
	for _, test := range messageDeleteTests {
		message, _ := New(
			SetData(test.data),
		)

		if err := message.Delete(test.key); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(message.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, message.Data())
		}
	}
}

func benchmarkTestMessageDeleteData(b *testing.B, key string, data []byte) {
	for i := 0; i < b.N; i++ {
		message, _ := New(
			SetData(data),
		)

		_ = message.Delete(key)
	}
}

func BenchmarkTestMessageDeleteData(b *testing.B) {
	for _, test := range messageDeleteTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTestMessageDeleteData(b, test.key, test.data)
			},
		)
	}
}

func TestMessageDeleteMetadata(t *testing.T) {
	for _, test := range messageDeleteTests {
		message, _ := New(
			SetMetadata(test.data),
		)

		key := strings.Join([]string{metaKey, test.key}, " ")
		if err := message.Delete(key); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(message.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, message.Metadata())
		}
	}
}

func benchmarkTestMessageDeleteMetadata(b *testing.B, key string, metadata []byte) {
	for i := 0; i < b.N; i++ {
		message, _ := New(
			SetMetadata(metadata),
		)

		_ = message.Delete(key)
	}
}

func BenchmarkTestMessageDeleteMetadata(b *testing.B) {
	for _, test := range messageDeleteTests {
		b.Run(test.name,
			func(b *testing.B) {
				key := strings.Join([]string{metaKey, test.key}, " ")
				benchmarkTestMessageDeleteMetadata(b, key, test.data)
			},
		)
	}
}

var messageGetTests = []struct {
	name     string
	data     []byte
	expected string
	key      string
}{
	{
		"a",
		[]byte(`{"a":"b","c":"d"}`),
		"b",
		"a",
	},
	{
		"@this",
		[]byte(`{"a":"b","c":"d"}`),
		`{"a":"b","c":"d"}`,
		"@this",
	},
	{
		"empty",
		[]byte(`{"a":"b","c":"d"}`),
		"",
		"",
	},
}

func TestMessageGetData(t *testing.T) {
	for _, test := range messageGetTests {
		message, _ := New(
			SetData(test.data),
		)

		result := message.Get(test.key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestMessageGetData(b *testing.B, key string, data []byte) {
	for i := 0; i < b.N; i++ {
		message, _ := New(
			SetData(data),
		)

		message.Get(key)
	}
}

func BenchmarkTestMessageGetData(b *testing.B) {
	for _, test := range messageGetTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTestMessageGetData(b, test.key, test.data)
			},
		)
	}
}

func TestMessageGetMetadata(t *testing.T) {
	for _, test := range messageGetTests {
		message, _ := New(
			SetMetadata(test.data),
		)

		key := strings.Join([]string{metaKey, test.key}, " ")
		result := message.Get(key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestMessageGetMetadata(b *testing.B, key string, metadata []byte) {
	for i := 0; i < b.N; i++ {
		message, _ := New(
			SetMetadata(metadata),
		)

		message.Get(key)
	}
}

func BenchmarkTestMessageGetMetadata(b *testing.B) {
	for _, test := range messageGetTests {
		b.Run(test.name,
			func(b *testing.B) {
				key := strings.Join([]string{metaKey, test.key}, " ")
				benchmarkTestMessageGetMetadata(b, key, test.data)
			},
		)
	}
}

var messageSetTests = []struct {
	name     string
	data     []byte
	expected []byte
	key      string
	value    interface{}
}{
	{
		"string",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d","e":"f"}`),
		"e",
		"f",
	},
	{
		"int",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d","e":1}`),
		"e",
		1,
	},
	{
		"float",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d","e":1.1}`),
		"e",
		1.1,
	},
	{
		"object",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d","e":{"f":"g"}}`),
		"e",
		[]byte(`{"f":"g"}`),
	},
	{
		"array",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d","e":["f","g","h"]}`),
		"e",
		[]byte(`["f","g","h"]`),
	},
	{
		"struct",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d","e":{"f":"g"}}`),
		"e",
		struct {
			F string `json:"f"`
		}{
			F: "g",
		},
	},
}

func TestMessageSetData(t *testing.T) {
	for _, test := range messageSetTests {
		message, _ := New(
			SetData(test.data),
		)

		if err := message.Set(test.key, test.value); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(message.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, message.Data())
		}
	}
}

func benchmarkTestMessageSetData(b *testing.B, key string, val interface{}, data []byte) {
	for i := 0; i < b.N; i++ {
		message, _ := New(
			SetData(data),
		)

		_ = message.Set(key, val)
	}
}

func BenchmarkTestMessageSetData(b *testing.B) {
	for _, test := range messageSetTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTestMessageSetData(b, test.key, test.value, test.data)
			},
		)
	}
}

func TestMessageSetMetadata(t *testing.T) {
	for _, test := range messageSetTests {
		message, _ := New(
			SetMetadata(test.data),
		)

		key := strings.Join([]string{metaKey, test.key}, " ")
		if err := message.Set(key, test.value); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(message.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, message.Metadata())
		}
	}
}

func benchmarkTestMessageSetMetadata(b *testing.B, key string, val interface{}, metadata []byte) {
	for i := 0; i < b.N; i++ {
		message, _ := New(
			SetMetadata(metadata),
		)

		_ = message.Set(key, val)
	}
}

func BenchmarkTestMessageSetMetadata(b *testing.B) {
	for _, test := range messageSetTests {
		b.Run(test.name,
			func(b *testing.B) {
				key := strings.Join([]string{metaKey, test.key}, " ")
				benchmarkTestMessageSetMetadata(b, key, test.value, test.data)
			},
		)
	}
}
