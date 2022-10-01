package process

import (
	"bytes"
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var domainTests = []struct {
	name     string
	proc     Domain
	test     []byte
	expected []byte
	err      error
}{
	{
		"JSON tld",
		Domain{
			Options: DomainOptions{
				Function: "tld",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"bar.com"}`),
		[]byte(`{"foo":"com"}`),
		nil,
	},
	{
		"JSON domain",
		Domain{
			Options: DomainOptions{
				Function: "domain",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"www.example.com"}`),
		[]byte(`{"foo":"example.com"}`),
		nil,
	},
	{
		"JSON subdomain",
		Domain{
			Options: DomainOptions{
				Function: "subdomain",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"www.bar.com"}`),
		[]byte(`{"foo":"www"}`),
		nil,
	},
	// empty subdomain, returns empty
	{
		"JSON subdomain",
		Domain{
			Options: DomainOptions{
				Function: "subdomain",
			},
			InputKey:  "foo",
			OutputKey: "foo",
		},
		[]byte(`{"foo":"example.com"}`),
		[]byte(`{"foo":""}`),
		nil,
	},
	{
		"data",
		Domain{
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		[]byte(`www.bar.com`),
		[]byte(`www`),
		nil,
	},
}

func TestDomain(t *testing.T) {
	ctx := context.TODO()
	cap := config.NewCapsule()

	for _, test := range domainTests {
		cap.SetData(test.test)

		result, err := test.proc.Apply(ctx, cap)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if !bytes.Equal(result.Data(), test.expected) {
			t.Logf("expected %s, got %s", test.expected, result.Data())
			t.Fail()
		}
	}
}

func benchmarkDomain(b *testing.B, applicator Domain, test config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		applicator.Apply(ctx, test)
	}
}

func BenchmarkDomain(b *testing.B) {
	cap := config.NewCapsule()
	for _, test := range domainTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				cap.SetData(test.test)
				benchmarkDomain(b, test.proc, cap)
			},
		)
	}
}
