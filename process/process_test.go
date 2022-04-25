package process

import (
	"bytes"
	"context"
	"testing"
)

var processTests = []struct {
	conf     []Config
	test     []byte
	expected []byte
}{
	{
		[]Config{
			{
				Type: "insert",
				Settings: map[string]interface{}{
					"condition": struct {
						Operator string
					}{
						Operator: "all",
					},
					"options": struct {
						Value interface{}
					}{
						Value: "bar",
					},
					"output": struct {
						Key string
					}{
						Key: "foo",
					},
				},
			},
		},
		[]byte(`{"hello":"world"}`),
		[]byte(`{"hello":"world","foo":"bar"}`),
	},
	{
		[]Config{
			{
				Type: "gzip",
				Settings: map[string]interface{}{
					"condition": struct {
						Operator string
					}{
						Operator: "all",
					},
					"options": struct {
						Direction string
					}{
						Direction: "from",
					},
				},
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 170, 86, 202, 72, 205, 201, 201, 87, 178, 82, 74, 207, 207, 79, 73, 170, 76, 85, 170, 5, 4, 0, 0, 255, 255, 214, 182, 196, 150, 19, 0, 0, 0},
		[]byte(`{"hello":"goodbye"}`),
	},
	{
		[]Config{
			{
				Type: "base64",
				Settings: map[string]interface{}{
					"condition": struct {
						Operator string
					}{
						Operator: "all",
					},
					"options": struct {
						Direction string
						Alphabet  string
					}{
						Direction: "from",
						Alphabet:  "std",
					},
				},
			},
		},
		[]byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
		[]byte(`{"hello":"world"}`),
	},
}

func TestByterAll(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processTests {
		byters, err := MakeAllByters(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed, err := Byte(ctx, byters, test.test)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(processed, test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

func TestByterFactory(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processTests {
		conf := test.conf[0]
		byter, err := ByterFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed, err := byter.Byte(ctx, test.test)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		if c := bytes.Compare(processed, test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

func TestChannel(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processTests {
		ch := make(chan []byte, 1)
		ch <- test.test
		close(ch)

		channelers, err := MakeAllChannelers(test.conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		output, err := Channel(ctx, channelers, ch)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed := <-output
		if c := bytes.Compare(processed, test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}

func TestChannelFactory(t *testing.T) {
	ctx := context.TODO()

	for _, test := range processTests {
		ch := make(chan []byte, 1)
		ch <- test.test
		close(ch)

		conf := test.conf[0]
		channeler, err := ChannelerFactory(conf)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		output, err := channeler.Channel(ctx, ch)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		processed := <-output
		if c := bytes.Compare(processed, test.expected); c != 0 {
			t.Logf("expected %v, got %v", test.expected, processed)
			t.Fail()
		}
	}
}
