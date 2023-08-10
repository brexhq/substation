package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var ipTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected bool
}{
	{
		"json",
		config.Config{
			Settings: map[string]interface{}{
				"key":  "ip_address",
				"type": "private",
			},
		},
		[]byte(`{"ip_address":"192.168.1.2"}`),
		true,
	},
	{
		"valid",
		config.Config{
			Settings: map[string]interface{}{
				"type": "valid",
			},
		},
		[]byte("192.168.1.2"),
		true,
	},
	{
		"invalid",
		config.Config{
			Settings: map[string]interface{}{
				"type": "valid",
			},
		},
		[]byte("foo"),
		false,
	},
	{
		"multicast",
		config.Config{
			Settings: map[string]interface{}{
				"type": "multicast",
			},
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"multicast_link_local",
		config.Config{
			Settings: map[string]interface{}{
				"type": "multicast_link_local",
			},
		},
		[]byte("224.0.0.12"),
		true,
	},
	{
		"unicast_global",
		config.Config{
			Settings: map[string]interface{}{
				"type": "unicast_global",
			},
		},
		[]byte("8.8.8.8"),
		true,
	},
	{
		"private",
		config.Config{
			Settings: map[string]interface{}{
				"type": "private",
			},
		},
		[]byte("8.8.8.8"),
		false,
	},
	{
		"unicast_link_local",
		config.Config{
			Settings: map[string]interface{}{
				"type": "unicast_link_local",
			},
		},
		[]byte("169.254.255.255"),
		true,
	},
	{
		"loopback",
		config.Config{
			Settings: map[string]interface{}{
				"type": "loopback",
			},
		},
		[]byte("127.0.0.1"),
		true,
	},
	{
		"unspecified",
		config.Config{
			Settings: map[string]interface{}{
				"type": "unspecified",
			},
		},
		[]byte("0.0.0.0"),
		true,
	},
}

func TestIP(t *testing.T) {
	ctx := context.TODO()

	for _, test := range ipTests {
		t.Run(test.name, func(t *testing.T) {
			message, _ := mess.New(
				mess.SetData(test.test),
			)

			insp, err := newInspIP(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Inspect(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v, %v", test.expected, check, string(test.test))
			}
		})
	}
}

func benchmarkIPByte(b *testing.B, inspector *inspIP, message *mess.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, message)
	}
}

func BenchmarkIPByte(b *testing.B) {
	for _, test := range ipTests {
		insp, err := newInspIP(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message, _ := mess.New(
					mess.SetData(test.test),
				)
				benchmarkIPByte(b, insp, message)
			},
		)
	}
}
