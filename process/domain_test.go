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
	{
		"invalid settings",
		Domain{},
		[]byte{},
		[]byte{},
		ProcessorInvalidSettings,
	},
}

func TestDomain(t *testing.T) {
	ctx := context.TODO()
	for _, test := range domainTests {
		res, err := test.proc.Byte(ctx, test.test)
		if err != nil && errors.Is(err, test.err) {
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
