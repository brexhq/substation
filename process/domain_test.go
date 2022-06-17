package process

import (
	"bytes"
	"context"
	"testing"
)

var domainTests = []struct {
	name     string
	proc     Domain
	test     []byte
	expected []byte
}{
	{
		"json tld",
		Domain{
			Input:  "domain",
			Output: "tld",
			Options: DomainOptions{
				Function: "tld",
			},
		},
		[]byte(`{"domain":"example.com"}`),
		[]byte(`{"domain":"example.com","tld":"com"}`),
	},
	{
		"json domain",
		Domain{
			Input:  "domain",
			Output: "domain",
			Options: DomainOptions{
				Function: "domain",
			},
		},
		[]byte(`{"domain":"www.example.com"}`),
		[]byte(`{"domain":"example.com"}`),
	},
	{
		"json subdomain",
		Domain{
			Input:  "domain",
			Output: "subdomain",
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		[]byte(`{"domain":"www.example.com"}`),
		[]byte(`{"domain":"www.example.com","subdomain":"www"}`),
	},
	// empty subdomain, returns empty
	{
		"json subdomain",
		Domain{
			Input:  "domain",
			Output: "subdomain",
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		[]byte(`{"domain":"example.com"}`),
		[]byte(`{"domain":"example.com","subdomain":""}`),
	},
	// array support
	{
		"json array tld",
		Domain{
			Input:  "domain",
			Output: "tld",
			Options: DomainOptions{
				Function: "tld",
			},
		},
		[]byte(`{"domain":["example.com","example.top"]}`),
		[]byte(`{"domain":["example.com","example.top"],"tld":["com","top"]}`),
	},
	{
		"json array domain",
		Domain{
			Input:  "domain",
			Output: "domain",
			Options: DomainOptions{
				Function: "domain",
			},
		},
		[]byte(`{"domain":["www.example.com","mail.example.top"]}`),
		[]byte(`{"domain":["example.com","example.top"]}`),
	},
	{
		"json array subdomain",
		Domain{
			Input:  "domain",
			Output: "subdomain",
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		[]byte(`{"domain":["www.example.com","mail.example.top"]}`),
		[]byte(`{"domain":["www.example.com","mail.example.top"],"subdomain":["www","mail"]}`),
	},
	{
		"data",
		Domain{
			Options: DomainOptions{
				Function: "subdomain",
			},
		},
		[]byte(`www.example.com`),
		[]byte(`www`),
	},
}

func TestDomain(t *testing.T) {
	for _, test := range domainTests {
		ctx := context.TODO()
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil {
			t.Logf("%v", err)
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
