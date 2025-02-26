package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"
)

var _ Conditioner = &stringEqualTo{}

var stringEqualToTests = []struct {
	name     string
	cfg      config.Config
	data     []byte
	expected bool
}{
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": "abcde",
			},
		},
		[]byte("abcde"),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"value": "abcde",
			},
		},
		[]byte("abcdef"),
		false,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"value": `""`,
			},
		},
		[]byte("\"\""),
		true,
	},
	{
		"pass",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "bar",
				},
			},
		},
		[]byte(`{"foo":"abc", "bar":"abc"}`),
		true,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "bar",
				},
				"value": "abc",
			},
		},
		[]byte(`{"foo":"abc", "bar":"def"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": "abc",
			},
		},
		[]byte(`{"bar": "abc", "baz": "0"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
				},
				"value": "",
			},
		},
		[]byte(`{"bar": "abc", "baz": "0"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "foo",
					"target_key": "baz",
				},
			},
		},
		[]byte(`{"bar": "abc", "baz": "0"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"target_key": "foo",
				},
			},
		},
		[]byte(`{"bar": "abc", "baz": "0"}`),
		false,
	},
	{
		"fail",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"source_key": "bar",
					"target_key": "foo",
				},
			},
		},
		[]byte(`{"bar": "abc", "baz": "0"}`),
		false,
	},
}

func TestStringEqualTo(t *testing.T) {
	ctx := context.TODO()

	for _, test := range stringEqualToTests {
		t.Run(test.name, func(t *testing.T) {
			message := message.New().SetData(test.data)

			insp, err := newStringEqualTo(ctx, test.cfg)
			if err != nil {
				t.Fatal(err)
			}

			check, err := insp.Condition(ctx, message)
			if err != nil {
				t.Error(err)
			}

			if test.expected != check {
				t.Errorf("expected %v, got %v", test.expected, check)
			}
		})
	}
}

func benchmarkStringEqualTo(b *testing.B, insp *stringEqualTo, message *message.Message) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = insp.Condition(ctx, message)
	}
}

func BenchmarkStringEqualTo(b *testing.B) {
	for _, test := range stringEqualToTests {
		insp, err := newStringEqualTo(context.TODO(), test.cfg)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(test.name,
			func(b *testing.B) {
				message := message.New().SetData(test.data)
				benchmarkStringEqualTo(b, insp, message)
			},
		)
	}
}

func FuzzTestStringEqualTo(f *testing.F) {
	testcases := [][]byte{
		[]byte(`"abcde"`),
		[]byte(`"abcdef"`),
		[]byte(`""`),
		[]byte(`"abcd"`),
		[]byte(`"12345"`),
		[]byte(`"abc"`),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx := context.TODO()
		message := message.New().SetData(data)
		insp, err := newStringEqualTo(ctx, config.Config{
			Settings: map[string]interface{}{
				"value": "abcde",
			},
		})
		if err != nil {
			return
		}

		_, err = insp.Condition(ctx, message)
		if err != nil {
			return
		}
	})
}
