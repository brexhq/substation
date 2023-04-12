package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var _ Inspector = inspContent{}

var contentTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	// matching Gzip against valid Gzip header
	{
		"gzip",
		config.Config{
			Type: "content",
			Settings: map[string]interface{}{
				"key": "ip_address",
				"options": map[string]interface{}{
					"type": "application/x-gzip",
				},
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		true,
	},
	// matching Gzip against invalid Gzip header (bytes swapped)
	{
		"!gzip",
		config.Config{
			Type: "content",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type": "application/x-gzip",
				},
			},
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		false,
	},
	// matching Gzip against invalid Gzip header (bytes swapped) with negation
	{
		"gzip",
		config.Config{
			Type: "content",
			Settings: map[string]interface{}{
				"negate": true,
				"options": map[string]interface{}{
					"type": "application/x-gzip",
				},
			},
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		true,
	},
	// matching Zip against valid Zip header
	{
		"zip",
		config.Config{
			Type: "content",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type": "application/zip",
				},
			},
		},
		[]byte{80, 75, 0o3, 0o4},
		true,
	},
	// matching Gzip against valid Zip header
	{
		"!zip",
		config.Config{
			Type: "content",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type": "application/zip",
				},
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		false,
	},
	// matching Zip against invalid Zip header (bytes swapped)
	{
		"!zip",
		// inspContent{
		// 	Options: inspContentOptions{
		// 		Type: "application/zip",
		// 	},
		// },
		config.Config{
			Type: "content",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type": "application/zip",
				},
			},
		},
		[]byte{0o4, 75, 0o3, 80},
		false,
	},
}

func TestContent(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range contentTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			insp, err := newInspContent(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkContent(b *testing.B, inspector inspContent, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkContent(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range contentTests {
		insp, err := newInspContent(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkContent(b, insp, capsule)
			},
		)
	}
}
