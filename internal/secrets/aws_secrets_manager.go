package secrets

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/brexhq/substation/v2/config"

	"github.com/brexhq/substation/v2/internal/aws"
	iconfig "github.com/brexhq/substation/v2/internal/config"
)

type awsSecretsManagerConfig struct {
	ID        string      `json:"id"`
	TTLOffset string      `json:"ttl_offset"`
	AWS       iconfig.AWS `json:"aws"`
}

func (c *awsSecretsManagerConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *awsSecretsManagerConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("id: %v", iconfig.ErrMissingRequiredOption)
	}

	if c.AWS.ARN == "" {
		return fmt.Errorf("aws.arn: %v", iconfig.ErrMissingRequiredOption)
	}

	return nil
}

type awsSecretsManager struct {
	conf   awsSecretsManagerConfig
	client *secretsmanager.Client

	ttl int64
}

func newAWSSecretsManager(ctx context.Context, cfg config.Config) (*awsSecretsManager, error) {
	conf := awsSecretsManagerConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	ttl := conf.TTLOffset
	if ttl == "" {
		ttl = defaultTTL
	}

	dur, err := time.ParseDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	c := &awsSecretsManager{
		conf: conf,
		ttl:  time.Now().Add(dur).Unix(),
	}

	awsCfg, err := aws.New(ctx, aws.Config{
		Region:  aws.ParseRegion(conf.AWS.ARN),
		RoleARN: conf.AWS.AssumeRoleARN,
	})
	if err != nil {
		return nil, fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	c.client = secretsmanager.NewFromConfig(awsCfg)

	return c, nil
}

func (c *awsSecretsManager) Retrieve(ctx context.Context) error {
	ctx = context.WithoutCancel(ctx)
	v, err := c.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &c.conf.AWS.ARN,
	})
	if err != nil {
		return fmt.Errorf("secrets: aws_secrets_manager: %v", err)
	}

	// SetWithTTL isn't used here because the TTL is managed by
	// transform/utility_secret.go.
	if err := cache.Set(ctx, c.conf.ID, v.SecretString); err != nil {
		return fmt.Errorf("secrets: aws_secrets_manager: id %s: %v", c.conf.ID, err)
	}

	return nil
}

func (c *awsSecretsManager) Expired() bool {
	return time.Now().Unix() >= c.ttl
}
