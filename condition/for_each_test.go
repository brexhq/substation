package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/brexhq/substation/config"
)

var forEachTests = []struct {
	name      string
	inspector ForEach
	test      []byte
	expected  bool
	err       error
}{
	{
		"strings startswith all",
		ForEach{
			Key:    "input",
			Negate: false,
			Mode:   "all",
			Options: ForEachOptions{
				Inspector: config.Config{
					Type: "strings",
					Settings: map[string]interface{}{
						"function":   "startswith",
						"expression": "f",
					},
				},
			},
		},
		[]byte(`{"input":["foo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"ip private all",
		ForEach{
			Key:    "input",
			Negate: false,
			Mode:   "all",
			Options: ForEachOptions{
				Inspector: config.Config{
					Type: "ip",
					Settings: map[string]interface{}{
						"type": "private",
					},
				},
			},
		},
		[]byte(`{"input":["192.168.1.2","10.0.42.1","172.16.4.2"]}`),
		true,
		nil,
	},
	{
		"regexp any",
		ForEach{
			Key:    "input",
			Negate: false,
			Mode:   "any",
			Options: ForEachOptions{
				Inspector: config.Config{
					Type: "regexp",
					Settings: map[string]interface{}{
						"expression": "^fizz$",
					},
				},
			},
		},
		[]byte(`{"input":["foo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"length none",
		ForEach{
			Key:    "input",
			Negate: false,
			Mode:   "none",
			Options: ForEachOptions{
				Inspector: config.Config{
					Type: "length",
					Settings: map[string]interface{}{
						"function": "greaterthan",
						"value":    7,
					},
				},
			},
		},
		[]byte(`{"input":["fooo","fizz","flop"]}`),
		true,
		nil,
	},
	{
		"length all",
		ForEach{
			Key:    "input",
			Negate: false,
			Mode:   "all",
			Options: ForEachOptions{
				Inspector: config.Config{
					Type: "length",
					Settings: map[string]interface{}{
						"function": "equals",
						"value":    4,
					},
				},
			},
		},
		[]byte(`{"input":["fooo","fizz","flop"]}`),
		true,
		nil,
	},
}

func TestForEach(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, tt := range forEachTests {
		t.Run(tt.name, func(t *testing.T) {
			capsule.SetData(tt.test)

			out, _ := json.Marshal(tt.inspector)
			fmt.Println(string(out))

			check, err := tt.inspector.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if tt.expected != check {
				t.Errorf("expected %v, got %v, %v", tt.expected, check, string(tt.test))
			}
		})
	}
}

func benchmarkForEachByte(b *testing.B, inspector ForEach, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkForEachByte(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range forEachTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkForEachByte(b, test.inspector, capsule)
			},
		)
	}
}
