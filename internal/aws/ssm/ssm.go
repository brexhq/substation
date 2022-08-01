package ssm

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// New returns a configured Systems Manager client.
func New() *ssm.SSM {
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

	c := ssm.New(
		session.Must(session.NewSession()),
		conf,
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the Systems Manager API interface.
type API struct {
	Client ssmiface.SSMAPI
}

// Setup creates a new Systems Manager client.
func (a *API) Setup() {
	a.Client = New()
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// GetParameter is a convenience wrapper for getting a parameter from Systems Manager.
func (a *API) GetParameter(ctx aws.Context, param string) (val string, err error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(param),
		WithDecryption: aws.Bool(true),
	}
	result, err := a.Client.GetParameterWithContext(ctx, input)
	if err != nil {
		return val, fmt.Errorf("getparameter parameter %s: %w", param, err)
	}

	val = *result.Parameter.Value
	return val, err
}
