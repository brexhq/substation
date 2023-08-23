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
	mess "github.com/brexhq/substation/message"
)

var _ Transformer = &procAWSLambda{}

type mockedInvoke struct {
	lambdaiface.LambdaAPI
	Resp lambda.InvokeOutput
}

func (m mockedInvoke) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	return &m.Resp, nil
}

var procAWSLambdaTests = []struct {
	name     string
	cfg      config.Config
	test     []byte
	expected [][]byte
	err      error
	api      lamb.API
}{
	{
		"JSON",
		config.Config{
			Settings: map[string]interface{}{
				"key":           "foo",
				"set_key":       "foo",
				"function_name": "fooer",
			},
		},
		[]byte(`{"foo":{"bar":"baz"}}`),
		[][]byte{
			[]byte(`{"foo":{"baz":"qux"}}`),
		},
		nil,
		lamb.API{
			Client: mockedInvoke{
				Resp: lambda.InvokeOutput{
					Payload: []byte(`{"baz":"qux"}`),
				},
			},
		},
	},
}

func TestprocAWSLambda(t *testing.T) {
	ctx := context.TODO()
	for _, test := range procAWSLambdaTests {
		message, err := mess.New(
			mess.SetData(test.test),
		)
		if err != nil {
			t.Fatal(err)
		}

		proc, err := newProcAWSLambda(ctx, test.cfg)
		if err != nil {
			t.Fatal(err)
		}
		proc.client = test.api

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
	}
}

func benchmarkprocAWSLambda(b *testing.B, tf *procAWSLambda, data []byte) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		message, _ := mess.New(
			mess.SetData(data),
		)

		_, _ = tf.Transform(ctx, message)
	}
}

func BenchmarkprocAWSLambda(b *testing.B) {
	ctx := context.TODO()
	for _, test := range procAWSLambdaTests {
		b.Run(test.name,
			func(b *testing.B) {
				proc, err := newProcAWSLambda(ctx, test.cfg)
				if err != nil {
					b.Fatal(err)
				}

				benchmarkprocAWSLambda(b, proc, test.test)
			},
		)
	}
}
