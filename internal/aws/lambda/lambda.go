package lambda

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

//New creates a new session connection to Lambda
func New() *lambda.Lambda {
	conf := aws.NewConfig()

	// provides forward compatibility for the Go SDK to support env var configuration settings
	// https://github.com/aws/aws-sdk-go/issues/4207
	max, found := os.LookupEnv("AWS_MAX_ATTEMPTS")
	if found {
		m, err := strconv.Atoi(max)
		if err != nil {
			panic(err)
		}

		conf = conf.WithMaxRetries(m)
	}

	c := lambda.New(
		session.Must(session.NewSession()),
		conf,
	)
	xray.AWS(c.Client)
	return c
}

// API wraps a lambda client interface
type API struct {
	Client lambdaiface.LambdaAPI
}

// Setup creates and sets a lambda client
func (a *API) Setup() {
	a.Client = New()
}

//IsEnabled returns the boolean on whether the client is enabled
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

//Invoke is a convenience wrapper around the AWS Lambda Invoke call with triggers a lambda
func (a *API) Invoke(ctx aws.Context, function string, payload []byte) (resp *lambda.InvokeOutput, err error) {
	resp, err = a.Client.InvokeWithContext(
		ctx,
		&lambda.InvokeInput{
			FunctionName:   aws.String(function),
			InvocationType: aws.String("RequestResponse"),
			Payload:        payload,
		})

	if err != nil {
		return resp, err
	}

	return resp, nil
}
