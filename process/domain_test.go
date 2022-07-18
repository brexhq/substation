package process

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

var domainTests = []struct {
	name     string
	proc     Domain
	err      error
	test     []byte
	expected []byte
}{
	{
		"json tld",
		Domain{
			InputKey:  "domain",
			OutputKey: "tld",
			Options: DomainOptions{
				Function: "tld",
			},
		},
		nil,
		[]byte(`{"domain":"example.com"}`),
		[]byte(`{"domain":"example.com","tld":"com"}`),
	},
	{
		"json domain",
		Domain{
			InputKey:  "domain",
			OutputKey: "domain",
			Options: DomainOptions{
				Function: "domain",
			},
		},
		nil,
		[]byte(`{"domain":"www.example.com"}`),
		[]byte(`{"domain":"example.com"}`),
	},
	{
		"json subdomain",
		Domain{
			InputKey:  "domain",
			OutputKey: "subdomain",
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		nil,
		[]byte(`{"domain":"www.example.com"}`),
		[]byte(`{"domain":"www.example.com","subdomain":"www"}`),
	},
	// empty subdomain, returns empty
	{
		"json subdomain",
		Domain{
			InputKey:  "domain",
			OutputKey: "subdomain",
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		nil,
		[]byte(`{"domain":"example.com"}`),
		[]byte(`{"domain":"example.com","subdomain":""}`),
	},
	{
		"data",
		Domain{
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		nil,
		[]byte(`www.example.com`),
		[]byte(`www`),
	},
	{
		"invalid settings",
		Domain{},
		ProcessorInvalidSettings,
		[]byte{},
		[]byte{},
	},
}

func TestDomain(t *testing.T) {
	for _, test := range domainTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.As(err, &test.err) {
			continue
		} else if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(res, test.expected); c != 0 {
			t.Logf("expected %s, got %s", test.expected, res)
			t.Fail()
		}
	}
}

func benchmarkDomainByte(b *testing.B, byter Domain, test []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		byter.Byte(ctx, test)
	}
}

func BenchmarkDomainByte(b *testing.B) {
	for _, test := range domainTests {
		b.Run(string(test.name),
			func(b *testing.B) {
				benchmarkDomainByte(b, test.proc, test.test)
			},
		)
	}
}
