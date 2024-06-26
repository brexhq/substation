package lambda

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	iaws "github.com/brexhq/substation/internal/aws"
)

// New returns a configured Lambda client.

func New(cfg iaws.Config) *lambda.Lambda {
	conf, sess := iaws.New(cfg)

	c := lambda.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the Lambda API interface.
type API struct {
	Client lambdaiface.LambdaAPI
}

// Setup creates a new Lambda client.
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// Invoke is a convenience wrapper for synchronously invoking a Lambda function.
func (a *API) Invoke(ctx aws.Context, function string, payload []byte) (resp *lambda.InvokeOutput, err error) {
	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.InvokeWithContext(
		ctx,
		&lambda.InvokeInput{
			FunctionName:   aws.String(function),
			InvocationType: aws.String("RequestResponse"),
			Payload:        payload,
		})
	if err != nil {
		return nil, fmt.Errorf("invoke function %s: %v", function, err)
	}

	return resp, nil
}

// InvokeAsync is a convenience wrapper for asynchronously invoking a Lambda function.
func (a *API) InvokeAsync(ctx aws.Context, function string, payload []byte) (resp *lambda.InvokeOutput, err error) {
	ctx = context.WithoutCancel(ctx)
	resp, err = a.Client.InvokeWithContext(
		ctx,
		&lambda.InvokeInput{
			FunctionName:   aws.String(function),
			InvocationType: aws.String("Event"),
			Payload:        payload,
		})
	if err != nil {
		return nil, fmt.Errorf("invoke_async function %s: %v", function, err)
	}

	return resp, nil
}
