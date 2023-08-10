package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/brexhq/substation/config"
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procDomain{}

var procDomainTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
}{
	{
		"JSON tld",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "tld",
			},
		},
		[]byte(`{"foo":"bar.com"}`),
		[][]byte{
			[]byte(`{"foo":"com"}`),
		},
		nil,
	},
	{
		"JSON domain",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "domain",
			},
		},
		[]byte(`{"foo":"www.example.com"}`),
		[][]byte{
			[]byte(`{"foo":"example.com"}`),
		},
		nil,
	},
	{
		"JSON subdomain",
		config.Config{
			Settings: map[string]interface{}{
				"key":     "foo",
				"set_key": "foo",
				"type":    "subdomain",
			},
		},
		[]byte(`{"foo":"www.bar.com"}`),
		[][]byte{
			[]byte(`{"foo":"www"}`),
		},
		nil,
	},
	// empty subdomain
	{
		"JSON subdomain",
		config.Config{
			Settings: map[string]interface{}{
				"key":              "foo",
				"set_key":          "foo",
				"type":             "subdomain",
				"error_on_failure": false,
			},
		},
		[]byte(`{"foo":"example.com"}`),
		[][]byte{
			[]byte(`{"foo":""}`),
		},
		nil,
	},
	{
		"data",
		config.Config{
			Settings: map[string]interface{}{
				"type": "subdomain",
			},
		},
		[]byte(`www.bar.com`),
		[][]byte{
			[]byte(`www`),
		},
		nil,
	},
}

func TestProcDomain(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procDomainTests {
		t.Run(test.name, func(t *testing.T) {
			message, err := mess.New(
				mess.SetData(test.test),
			)
			if err != nil {
				t.Fatal(err)
			}

			proc, err := newProcDomain(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			result, err := proc.Transform(ctx, message)
			if err != nil {
				t.Error(err)
			}

			var data [][]byte
			for _, c := range result {
				data = append(data, c.Data())
			}

			if !reflect.DeepEqual(data, test.expected) {
				t.Errorf("expected %s, got %s", test.expected, data)
			}
		})
	}
}

func benchmarkProcDomain(b *testing.B, tform *procDomain, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tform.Transform(ctx, message)
	}
}

func BenchmarkProcDomain(b *testing.B) {
	for _, test := range procDomainTests {
		proc, err := newProcDomain(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				benchmarkProcDomain(b, proc, test.test)
			},
		)
	}
}
