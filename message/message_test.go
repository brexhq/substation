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
		msg := New().SetData(test.data)

		if err := msg.DeleteObject(test.key); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(msg.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Data())
		}
	}
}

func benchmarkTestMessageDeleteData(b *testing.B, key string, data []byte) {
	for i := 0; i < b.N; i++ {
		message := New().SetData(data)
		_ = message.DeleteObject(key)
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
		message := New().SetMetadata(test.data)

		key := strings.Join([]string{metaKey, test.key}, " ")
		if err := message.DeleteObject(key); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(message.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, message.Metadata())
		}
	}
}

func benchmarkTestMessageDeleteMetadata(b *testing.B, key string, metadata []byte) {
	for i := 0; i < b.N; i++ {
		message := New().SetMetadata(metadata)

		_ = message.DeleteObject(key)
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
		msg := New().SetData(test.data)

		result := msg.GetObject(test.key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestMessageGetData(b *testing.B, key string, data []byte) {
	for i := 0; i < b.N; i++ {
		msg := New().SetData(data)
		msg.GetObject(key)
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
		msg := New().SetMetadata(test.data)

		key := strings.Join([]string{metaKey, test.key}, " ")
		result := msg.GetObject(key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestMessageGetMetadata(b *testing.B, key string, metadata []byte) {
	for i := 0; i < b.N; i++ {
		msg := New().SetMetadata(metadata)
		msg.GetObject(key)
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
		msg := New().SetData(test.data)
		if err := msg.SetObject(test.key, test.value); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(msg.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Data())
		}
	}
}

func benchmarkTestMessageSetData(b *testing.B, key string, val interface{}, data []byte) {
	for i := 0; i < b.N; i++ {
		msg := New().SetData(data)
		_ = msg.SetObject(key, val)
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
		msg := New().SetMetadata(test.data)

		key := strings.Join([]string{metaKey, test.key}, " ")
		if err := msg.SetObject(key, test.value); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(msg.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Metadata())
		}
	}
}

func benchmarkTestMessageSetMetadata(b *testing.B, key string, val interface{}, metadata []byte) {
	for i := 0; i < b.N; i++ {
		msg := New().SetMetadata(metadata)
		_ = msg.SetObject(key, val)
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
