package secretsmanager

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// New returns a configured Secrets Manager client.
func New() *secretsmanager.SecretsManager {
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

	c := secretsmanager.New(
		session.Must(session.NewSession()),
		conf,
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the Secrets Manager API interface.
type API struct {
	Client secretsmanageriface.SecretsManagerAPI
}

// Setup creates a new Secrets Manager client.
func (a *API) Setup() {
	a.Client = New()
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// GetSecret is a convenience wrapper for getting a secret from Secrets Manager.
func (a *API) GetSecret(ctx aws.Context, secretName string) (secret string, err error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := a.Client.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return secret, fmt.Errorf("getsecretvalue secret %s: %w", secretName, err)
	}

	if result.SecretString != nil {
		secret = *result.SecretString
		return secret, err
	}

	return secret, err
}
