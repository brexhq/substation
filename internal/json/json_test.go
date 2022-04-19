package json

import (
	"testing"
)

var getTests = []struct {
	json     string
	key      string
	expected interface{}
}{
	{
		json:     `{"get":"string"}`,
		key:      "get",
		expected: "string",
	},
	{
		json:     `{"get":123}`,
		key:      "get",
		expected: 123.0,
	},
	{
		json:     `{"get":123.456}`,
		key:      "get",
		expected: 123.456,
	},
	{
		json:     `{"get":true}`,
		key:      "get",
		expected: true,
	},
	{
		json:     `{"get":false}`,
		key:      "get",
		expected: false,
	},
}

func TestGetJson(t *testing.T) {
	for _, gt := range getTests {
		result := Get([]byte(gt.json), gt.key)

		if result.Value() != gt.expected {
			t.Logf("expected %v, got %v", gt.expected, result)
			t.Fail()
		}
	}
}

var setTests = []struct {
	json     string
	key      string
	value    interface{}
	expected string
}{
	{
		json:     `{"hello":"world"}`,
		key:      "olleh",
		value:    "dlrow",
		expected: `{"hello":"world","olleh":"dlrow"}`,
	},
	{
		json:     `{"hello":"world"}`,
		key:      "olleh",
		value:    123,
		expected: `{"hello":"world","olleh":123}`,
	},
	{
		json:     `{"hello":"world"}`,
		key:      "olleh",
		value:    123.456,
		expected: `{"hello":"world","olleh":123.456}`,
	},
	{
		json:     `{"hello":"world"}`,
		key:      "olleh",
		value:    true,
		expected: `{"hello":"world","olleh":true}`,
	},
}

func TestSetJson(t *testing.T) {
	for _, st := range setTests {
		result, err := Set([]byte(st.json), st.key, st.value)
		if err != nil {
			t.Logf("got error %v", err)
			t.Fail()
			return
		}

		if string(result) != st.expected {
			t.Logf("expected %s, got %s", st.expected, result)
			t.Fail()
		}
	}
}

var setRawTests = []struct {
	json     string
	key      string
	value    string
	expected string
}{
	{
		json:     `{"hello":"world"}`,
		key:      "inner",
		value:    `{"olleh":"dlrow"}`,
		expected: `{"hello":"world","inner":{"olleh":"dlrow"}}`,
	},
}

func TestSetRawJson(t *testing.T) {
	for _, st := range setRawTests {
		result, err := SetRaw([]byte(st.json), st.key, st.value)
		if err != nil {
			t.Logf("got error %v", err)
			t.Fail()
			return
		}

		if string(result) != st.expected {
			t.Logf("expected %s, got %s", st.expected, result)
			t.Fail()
		}
	}
}

var validTests = []struct {
	json     string
	expected bool
}{
	{
		json:     `{"hello":"world"}`,
		expected: true,
	},
	{
		json:     `{hello:"world"}`,
		expected: false,
	},
}

func TestValidJson(t *testing.T) {
	for _, vt := range validTests {
		result := Valid([]byte(vt.json))

		if result != vt.expected {
			t.Logf("expected %v, got %v", vt.expected, result)
			t.Fail()
		}
	}
}
