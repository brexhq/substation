package process

import (
	"bytes"
	"context"
	"testing"
)

func TestDomain(t *testing.T) {
	var tests = []struct {
		proc     Domain
		test     []byte
		expected []byte
	}{
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "tld",
				},
				Output: Output{
					Key: "tld",
				},
			},
			[]byte(`{"domain":"example.com"}`),
			[]byte(`{"domain":"example.com","tld":"com"}`),
		},
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "domain",
				},
				Output: Output{
					Key: "domain",
				},
			},
			[]byte(`{"domain":"www.example.com"}`),
			[]byte(`{"domain":"example.com"}`),
		},
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "subdomain",
				},
				Output: Output{
					Key: "subdomain",
				},
			},
			[]byte(`{"domain":"www.example.com"}`),
			[]byte(`{"domain":"www.example.com","subdomain":"www"}`),
		},
		// empty subdomain, returns input
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "subdomain",
				},
				Output: Output{
					Key: "subdomain",
				},
			},
			[]byte(`{"domain":"example.com"}`),
			[]byte(`{"domain":"example.com"}`),
		},
		// array support
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "tld",
				},
				Output: Output{
					Key: "tld",
				},
			},
			[]byte(`{"domain":["example.com","example.top"]}`),
			[]byte(`{"domain":["example.com","example.top"],"tld":["com","top"]}`),
		},
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "domain",
				},
				Output: Output{
					Key: "domain",
				},
			},
			[]byte(`{"domain":["www.example.com","mail.example.top"]}`),
			[]byte(`{"domain":["example.com","example.top"]}`),
		},
		{
			Domain{
				Input: Input{
					Key: "domain",
				},
				Options: DomainOptions{
					Function: "subdomain",
				},
				Output: Output{
					Key: "subdomain",
				},
			},
			[]byte(`{"domain":["www.example.com","mail.example.top"]}`),
			[]byte(`{"domain":["www.example.com","mail.example.top"],"subdomain":["www","mail"]}`),
		},
	}

	for _, test := range tests {
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
