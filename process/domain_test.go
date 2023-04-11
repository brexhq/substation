package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var (
	_ Applier = procDomain{}
	_ Batcher = procDomain{}
)

var domainTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON tld",
		// procDomain{
		// 	process: process{
		// 		Key:    "foo",
		// 		SetKey: "foo",
		// 	},
		// 	Options: procDomainOptions{
		// 		Type: "tld",
		// 	},
		// },
		config.Config{
			Type: "domain",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "tld",
				},
			},
		},
		[]byte(`{"foo":"bar.com"}`),
		[]byte(`{"foo":"com"}`),
		nil,
	},
	{
		"JSON domain",
		// procDomain{
		// 	process: process{
		// 		Key:    "foo",
		// 		SetKey: "foo",
		// 	},
		// 	Options: procDomainOptions{
		// 		Type: "domain",
		// 	},
		// },
		config.Config{
			Type: "domain",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "domain",
				},
			},
		},
		[]byte(`{"foo":"www.example.com"}`),
		[]byte(`{"foo":"example.com"}`),
		nil,
	},
	{
		"JSON subdomain",
		// procDomain{
		// 	process: process{
		// 		Key:    "foo",
		// 		SetKey: "foo",
		// 	},
		// 	Options: procDomainOptions{
		// 		Type: "subdomain",
		// 	},
		// },
		config.Config{
			Type: "domain",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "subdomain",
				},
			},
		},
		[]byte(`{"foo":"www.bar.com"}`),
		[]byte(`{"foo":"www"}`),
		nil,
	},
	// empty subdomain, returns empty
	{
		"JSON subdomain",
		// procDomain{
		// 	process: process{
		// 		Key:    "foo",
		// 		SetKey: "foo",
		// 	},
		// 	Options: procDomainOptions{
		// 		Type: "subdomain",
		// 	},
		// },
		config.Config{
			Type: "domain",
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"options": map[string]interface{}{
					"type": "subdomain",
				},
			},
		},
		[]byte(`{"foo":"example.com"}`),
		[]byte(`{"foo":""}`),
		nil,
	},
	{
		"data",
		// procDomain{
		// 	Options: procDomainOptions{
		// 		Type: "subdomain",
		// 	},
		// },
		config.Config{
			Type: "domain",
			Settings: map[string]interface{}{
				"options": map[string]interface{}{
					"type": "subdomain",
				},
			},
		},
		[]byte(`www.bar.com`),
		[]byte(`www`),
		nil,
	},
}

func TestDomain(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range domainTests {
		t.Run(test.name, func(t *testing.T) {
			capsule.SetData(test.test)

			proc, err := newProcDomain(test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Apply(ctx, capsule)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result.Data(), test.expected) {
				t.Errorf("expected %s, got %s", test.expected, result.Data())
			}
		})
	}
}

func benchmarkDomain(b *testing.B, applier procDomain, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
	}
}

func BenchmarkDomain(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range domainTests {
		proc, err := newProcDomain(test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkDomain(b, proc, capsule)
			},
		)
	}
}
