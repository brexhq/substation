package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

var formatContentTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	// Matching Gzip against valid Gzip header.
	{
		"pass gzip",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "ip_address",
				},
				"type": "application/x-gzip",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		true,
	},
	// Matching Gzip against invalid Gzip header (bytes swapped).
	{
		"fail gzip",
		config.Config{
			Settings: map[string]interface{}{
				"type": "application/x-gzip",
			},
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		false,
	},
	// Matching Zip against valid Zip header.
	{
		"pass zip",
		config.Config{
			Settings: map[string]interface{}{
				"type": "application/zip",
			},
		},
		[]byte{80, 75, 0o3, 0o4},
		true,
	},
	// Matching Gzip against valid Zip header.
	{
		"fail zip",
		config.Config{
			Settings: map[string]interface{}{
				"type": "application/zip",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		false,
	},
	// Matching Zip against invalid Zip header (bytes swapped).
	{
		"fail zip",
		config.Config{
			Settings: map[string]interface{}{
				"type": "application/zip",
			},
		},
		[]byte{0o4, 75, 0o3, 80},
		false,
	},
}

func TestFormatContent(t *testing.T) {
	ctx := context.TODO()

	for _, test := range formatContentTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.test)
			insp, err := newFormatContent(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkFormatContent(b *testing.B, insp *formatContent, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Inspect(ctx, message)
	}
}

func BenchmarkFormatContent(b *testing.B) {
	for _, test := range formatContentTests {
		insp, err := newFormatContent(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.test)
				benchmarkFormatContent(b, insp, message)
			},
		)
	}
}
