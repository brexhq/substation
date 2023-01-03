package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var domainTests = []struct {
	name     string
	proc     procDomain
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON tld",
		procDomain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procDomainOptions{
				Type: "tld",
			},
		},
		[]byte(`{"foo":"bar.com"}`),
		[]byte(`{"foo":"com"}`),
		nil,
	},
	{
		"JSON procDomain",
		procDomain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procDomainOptions{
				Type: "domain",
			},
		},
		[]byte(`{"foo":"www.example.com"}`),
		[]byte(`{"foo":"example.com"}`),
		nil,
	},
	{
		"JSON subdomain",
		procDomain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procDomainOptions{
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
		procDomain{
			process: process{
				Key:    "foo",
				SetKey: "foo",
			},
			Options: procDomainOptions{
				Type: "subdomain",
			},
		},
		[]byte(`{"foo":"example.com"}`),
		[]byte(`{"foo":""}`),
		nil,
	},
	{
		"data",
		procDomain{
			Options: procDomainOptions{
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
		var _ Applier = test.proc
		var _ Batcher = test.proc

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

func benchmarkDomain(b *testing.B, applier procDomain, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = applier.Apply(ctx, test)
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
