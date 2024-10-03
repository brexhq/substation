package message

import (
	"bytes"
	"strings"
	"testing"
)

var messageNewTests = []struct {
	name     string
	data     []byte
	expected []byte
}{
	{
		"empty",
		[]byte{},
		[]byte{},
	},
	{
		"data",
		[]byte(`{"a":"b","c":"d"}`),
		[]byte(`{"a":"b","c":"d"}`),
	},
}

func TestMessageNew(t *testing.T) {
	for _, test := range messageNewTests {
		msg := New().SetData(test.data)

		if !bytes.Equal(msg.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Data())
		}
	}
}

func benchmarkTestMessageNew(b *testing.B, data []byte) {
	for i := 0; i < b.N; i++ {
		_ = New().SetData(data)
	}
}

func BenchmarkTestMessageNew(b *testing.B) {
	for _, test := range messageNewTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkTestMessageNew(b, test.data)
			},
		)
	}
}

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
}

func TestMessageDeleteData(t *testing.T) {
	for _, test := range messageDeleteTests {
		msg := New().SetData(test.data)

		if err := msg.DeleteValue(test.key); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(msg.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Data())
		}
	}
}

func benchmarkTestMessageDeleteData(b *testing.B, key string, data []byte) {
	msg := New().SetData(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = msg.DeleteValue(key)
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
		if err := message.DeleteValue(key); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(message.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, message.Metadata())
		}
	}
}

func benchmarkTestMessageDeleteMetadata(b *testing.B, key string, data []byte) {
	msg := New().SetMetadata(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = msg.DeleteValue(key)
	}
}

func BenchmarkTestMessageDeleteMetadata(b *testing.B) {
	for _, test := range messageDeleteTests {
		key := strings.Join([]string{metaKey, test.key}, " ")

		b.Run(test.name,
			func(b *testing.B) {
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

		result := msg.GetValue(test.key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestMessageGetData(b *testing.B, key string, data []byte) {
	msg := New().SetData(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = msg.GetValue(key)
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
		result := msg.GetValue(key).String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func benchmarkTestMessageGetMetadata(b *testing.B, key string, data []byte) {
	msg := New().SetMetadata(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = msg.GetValue(key)
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
		if err := msg.SetValue(test.key, test.value); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(msg.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Data())
		}
	}
}

func benchmarkTestMessageSetData(b *testing.B, key string, val interface{}, data []byte) {
	msg := New().SetData(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = msg.SetValue(key, val)
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
		if err := msg.SetValue(key, test.value); err != nil {
			t.Error(err)
		}

		if !bytes.Equal(msg.Metadata(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, msg.Metadata())
		}
	}
}

func benchmarkTestMessageSetMetadata(b *testing.B, key string, val interface{}, data []byte) {
	msg := New().SetMetadata(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = msg.SetValue(key, val)
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

func FuzzMessageSetValue(f *testing.F) {
	testcases := []struct {
		path  string
		value string
	}{
		{"a", "b"},
		{"a.b", "123"},
		{"a.b.c", "true"},
		{"", "empty path"},
		{"a.b.c.d", ""},
	}

	for _, tc := range testcases {
		f.Add(tc.path, tc.value)
	}

	f.Fuzz(func(t *testing.T, path string, value string) {
		msg := New()
		err := msg.SetValue(path, value)
		if err != nil {
			if err.Error() == "path cannot be empty" {
				return
			}
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func FuzzMessageGetValue(f *testing.F) {
	f.Add("key")
	f.Fuzz(func(t *testing.T, key string) {
		msg := New().SetData([]byte(`{"key":"value"}`))
		_ = msg.GetValue(key)
	})
}

func FuzzMessageDeleteValue(f *testing.F) {
	f.Add("key")
	f.Fuzz(func(t *testing.T, key string) {
		msg := New().SetData([]byte(`{"key":"value"}`))
		_ = msg.DeleteValue(key)
	})
}
