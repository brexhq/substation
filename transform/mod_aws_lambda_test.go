package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/brexhq/substation/config"
	lamb "github.com/brexhq/substation/internal/aws/lambda"
	"github.com/brexhq/substation/message"
)

var _ Transformer = &modAWSLambda{}

type mockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp lambda.InvokeOutput
}

func (m mockedInvoke) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	return &m.Resp, nil
}

var modAWSLambdaTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
	api      lamb.API
}{
	{
		"success",
		config.Config{
			Settings: map[string]interface{}{
				"object": map[string]interface{}{
					"key":     "a",
					"set_key": "a",
				},
				"function_name": "func",
			},
		},
		[]byte(`{"a":{"b":"c"}}`),
		[][]byte{
			[]byte(`{"a":{"d":"e"}}`),
		},
		nil,
		lamb.API{
			Client: mockedInvoke{
				Resp: lambda.InvokeOutput{
					Payload: []byte(`{"d":"e"}`),
				},
			},
		},
	},
}

func TestModAWSLambda(t *testing.T) {
	ctx := context.TODO()
	for _, test := range modAWSLambdaTests {
		tf, err := newModAWSLambda(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}
		tf.client = test.api

		msg := message.New().SetData(test.test)
		result, err := tf.Transform(ctx, msg)
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
	}
}

func benchmarkModAWSLambda(b *testing.B, tf *modAWSLambda, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		msg := message.New().SetData(data)
		_, _ = tf.Transform(ctx, msg)
	}
}

func BenchmarkModAWSLambda(b *testing.B) {
	ctx := context.TODO()
	for _, test := range modAWSLambdaTests {
		b.Run(test.name,
			func(b *testing.B) {
				tf, err := newModAWSLambda(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}
				tf.client = test.api

				benchmarkModAWSLambda(b, tf, test.test)
			},
		)
	}
}
