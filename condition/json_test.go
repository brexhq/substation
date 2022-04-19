package condition

import (
	"testing"
)

func TestJSONSchema(t *testing.T) {
	var tests = []struct {
		inspector JSONSchema
		test      []byte
		expected  bool
	}{
		{
			JSONSchema{
				Schema: []struct {
					Key  string `mapstructure:"key"`
					Type string `mapstructure:"type"`
				}{
					{Key: "hello", Type: "String"},
				},
				Negate: false,
			},
			[]byte(`{"hello":"world"}`),
			true,
		},
		{
			JSONSchema{
				Schema: []struct {
					Key  string `mapstructure:"key"`
					Type string `mapstructure:"type"`
				}{
					{Key: "hello", Type: "String"},
				},
				Negate: false,
			},
			[]byte(`{"hello":123}`),
			false,
		},
	}

	for _, testing := range tests {
		ok, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != ok {
			t.Logf("expected %v, got %v, %v", testing.expected, ok, string(testing.test))
			t.Fail()
		}
	}
}

func TestJSONValid(t *testing.T) {
	var tests = []struct {
		inspector JSONValid
		test      []byte
		expected  bool
	}{
		{
			JSONValid{},
			[]byte(`{"hello":"world"}`),
			true,
		},
		{
			JSONValid{},
			[]byte(`{hello:"world"}`),
			false,
		},
		{
			JSONValid{
				Negate: true,
			},
			[]byte(`{"hello":"world"}`),
			false,
		},
		{
			JSONValid{
				Negate: true,
			},
			[]byte(`{hello:"world"}`),
			true,
		},
	}

	for _, testing := range tests {
		ok, _ := testing.inspector.Inspect(testing.test)

		if testing.expected != ok {
			t.Logf("expected %v, got %v, %v", testing.expected, ok, string(testing.test))
			t.Fail()
		}
	}
}
