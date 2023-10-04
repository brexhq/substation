package secrets

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/secretsmanager"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
)

type awsSecretsManagerConfig struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	TTLOffset int64         `json:"ttl_offset"`
	AWS       iconfig.AWS   `json:"aws"`
	Retry     iconfig.Retry `json:"retry"`
}

func (c *awsSecretsManagerConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *awsSecretsManagerConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("id: %v", errors.ErrMissingRequiredOption)
	}

	if c.Name == "" {
		return fmt.Errorf("name: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

type awsSecretsManager struct {
	conf awsSecretsManagerConfig

	ttl int64
	// client is safe for concurrent access.
	client secretsmanager.API
}

func newAWSSecretsManager(_ context.Context, cfg config.Config) (*awsSecretsManager, error) {
	conf := awsSecretsManagerConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	ttl := conf.TTLOffset
	if ttl == 0 {
		ttl = defaultTTL
	}

	c := &awsSecretsManager{
		conf: conf,
		ttl:  time.Now().Add(time.Duration(ttl) * time.Second).Unix(),
	}

	c.client.Setup(aws.Config{
		Region:        conf.AWS.Region,
		AssumeRoleARN: conf.AWS.AssumeRoleARN,
		MaxRetries:    conf.Retry.Count,
	})

	return c, nil
}

func (c *awsSecretsManager) Collect(ctx context.Context) error {
	v, err := c.client.GetSecret(ctx, c.conf.Name)
	if err != nil {
		return fmt.Errorf("secrets: environment_variable: name %s: %v", c.conf.Name, err)
	}

	// SetWithTTL isn't used here because the TTL is managed by
	// transform/utility_secret.go.
	if err := cache.Set(ctx, c.conf.ID, v); err != nil {
		return fmt.Errorf("secrets: environment_variable: id %s: %v", c.conf.ID, err)
	}

	return nil
}

func (c *awsSecretsManager) Expired() bool {
	return time.Now().Unix() >= c.ttl
}
