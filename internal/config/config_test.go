package config

import (
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
