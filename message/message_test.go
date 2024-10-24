package message

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
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
		"missing",
		[]byte(`{"a":"b","c":"d"}`),
		"",
		"e",
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

var valueIsNullTests = []struct {
	name     string
	data     Value
	expected bool
}{
	// Only literal null is considered null.
	{
		"null",
		Value{gjson.Parse(`null`)},
		true,
	},
	// Empty values are not null.
	{
		"not null",
		Value{gjson.Parse(`""`)},
		false,
	},
}

func TestValueIsNull(t *testing.T) {
	for _, test := range valueIsNullTests {
		result := test.data.IsNull()
		if result != test.expected {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
	}
}

var valueIsMissingTests = []struct {
	name     string
	data     Value
	expected bool
}{
	// Only nil is considered missing.
	{
		"missing",
		Value{},
		true,
	},
	// Empty values are not missing.
	{
		"not missing",
		Value{gjson.Parse(`""`)},
		false,
	},
}

func TestValueIsMissing(t *testing.T) {
	for _, test := range valueIsMissingTests {
		result := test.data.IsMissing()
		if result != test.expected {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
	}
}

var valueIsEmptyTests = []struct {
	name     string
	data     Value
	expected bool
}{
	{
		"empty",
		Value{},
		true,
	},
	{
		"empty string",
		Value{gjson.Parse(`""`)},
		true,
	},
	{
		"empty object",
		Value{gjson.Parse(`{}`)},
		true,
	},
	{
		"empty array",
		Value{gjson.Parse(`[]`)},
		true,
	},
	{
		"null",
		Value{gjson.Parse(`null`)},
		true,
	},
	{
		"non-empty string",
		Value{gjson.Parse(`"foo"`)},
		false,
	},
	{
		"non-empty object",
		Value{gjson.Parse(`{"foo":"bar"}`)},
		false,
	},
	{
		"non-empty array",
		Value{gjson.Parse(`["foo","bar"]`)},
		false,
	},
}

func TestValueIsEmpty(t *testing.T) {
	for _, test := range valueIsEmptyTests {
		result := test.data.IsEmpty()
		if result != test.expected {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
	}
}

func benchmarkValueIsNull(b *testing.B, value Value) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = value.IsNull()
	}
}

func BenchmarkValueIsNull(b *testing.B) {
	for _, test := range valueIsNullTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkValueIsNull(b, test.data)
			},
		)
	}
}

func benchmarkValueIsMissing(b *testing.B, value Value) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = value.IsMissing()
	}
}

func BenchmarkValueIsMissing(b *testing.B) {
	for _, test := range valueIsMissingTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkValueIsMissing(b, test.data)
			},
		)
	}
}

func benchmarkValueIsEmpty(b *testing.B, value Value) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = value.IsEmpty()
	}
}

func BenchmarkValueIsEmpty(b *testing.B) {
	for _, test := range valueIsEmptyTests {
		b.Run(test.name,
			func(b *testing.B) {
				benchmarkValueIsEmpty(b, test.data)
			},
		)
	}
}

func FuzzValueIsNull(f *testing.F) {
	f.Add([]byte("null"))
	f.Add([]byte(`""`))
	f.Fuzz(func(t *testing.T, value []byte) {
		_ = Value{gjson.Parse(string(value))}.IsNull()
	})
}

func FuzzValueIsMissing(f *testing.F) {
	f.Add([]byte("null"))
	f.Add([]byte(`""`))
	f.Fuzz(func(t *testing.T, value []byte) {
		_ = Value{gjson.Parse(string(value))}.IsMissing()
	})
}

func FuzzValueIsEmpty(f *testing.F) {
	f.Add([]byte("null"))
	f.Add([]byte(`""`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Fuzz(func(t *testing.T, value []byte) {
		_ = Value{gjson.Parse(string(value))}.IsEmpty()
	})
}
