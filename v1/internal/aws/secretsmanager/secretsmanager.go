package secretsmanager

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-xray-sdk-go/xray"
	iaws "github.com/brexhq/substation/internal/aws"
)

// New returns a configured Secrets Manager client.
func New(cfg iaws.Config) *secretsmanager.SecretsManager {
	conf, sess := iaws.New(cfg)

	c := secretsmanager.New(sess, conf)
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
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
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

	ctx = context.WithoutCancel(ctx)
	result, err := a.Client.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return secret, fmt.Errorf("getsecretvalue secret %s: %v", secretName, err)
	}

	if result.SecretString != nil {
		secret = *result.SecretString
		return secret, err
	}

	return secret, err
}
