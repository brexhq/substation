package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var domainTests = []struct {
	name     string
	proc     domain
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON tld",
		domain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: domainOptions{
				Type: "tld",
			},
		},
		[]byte(`{"foo":"bar.com"}`),
		[]byte(`{"foo":"com"}`),
		nil,
	},
	{
		"JSON domain",
		domain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: domainOptions{
				Type: "domain",
			},
		},
		[]byte(`{"foo":"www.example.com"}`),
		[]byte(`{"foo":"example.com"}`),
		nil,
	},
	{
		"JSON subdomain",
		domain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: domainOptions{
				Type: "subdomain",
			},
		},
		[]byte(`{"foo":"www.bar.com"}`),
		[]byte(`{"foo":"www"}`),
		nil,
	},
	// empty subdomain, returns empty
	{
		"JSON subdomain",
		domain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: domainOptions{
				Type: "subdomain",
			},
		},
		[]byte(`{"foo":"example.com"}`),
		[]byte(`{"foo":""}`),
		nil,
	},
	{
		"data",
		domain{
			Options: domainOptions{
				Type: "subdomain",
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
		capsule.SetData(test.test)

		result, err := test.proc.Apply(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Errorf("expected %s, got %s", test.expected, result.Data())
		}
	}
}

func benchmarkDomain(b *testing.B, applicator domain, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applicator.Apply(ctx, test)
	}
}

func BenchmarkDomain(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range domainTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkDomain(b, test.proc, capsule)
			},
		)
	}
}
