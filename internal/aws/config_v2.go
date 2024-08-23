package aws

import (
	"context"
	"os"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

// NewV2 returns an SDK v2 configuration.
func NewV2(ctx context.Context, cfg Config) (aws.Config, error) {
	var region string
	if cfg.Region != "" {
		region = cfg.Region
	} else if v, ok := os.LookupEnv("AWS_REGION"); ok {
		region = v
	} else if v, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok {
		region = v
	}

	var creds aws.CredentialsProvider // nil is a valid default.
	if cfg.RoleARN != "" {
		conf, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
		if err != nil {
			return aws.Config{}, err
		}

		stsSvc := sts.NewFromConfig(conf)
		creds = stscreds.NewAssumeRoleProvider(stsSvc, cfg.RoleARN)
	}

	maxRetry := 3 // Matches the standard retryer.
	if cfg.MaxRetries != 0 {
		maxRetry = cfg.MaxRetries
	} else if v, ok := os.LookupEnv("AWS_MAX_ATTEMPTS"); ok {
		max, err := strconv.Atoi(v)
		if err != nil {
			return aws.Config{}, err
		}

		maxRetry = max
	}

	errMsg := make([]*regexp.Regexp, len(cfg.RetryableErrors))
	for i, err := range cfg.RetryableErrors {
		errMsg[i] = regexp.MustCompile(err)
	}

	conf, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.MaxAttempts = maxRetry
				// Additional retryable errors ~must be appended~ to not overwrite the defaults.
				o.Retryables = append(o.Retryables, retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
					for _, msg := range errMsg {
						if msg.MatchString(err.Error()) {
							return aws.TrueTernary
						}
					}

					return aws.FalseTernary
				}))
			})
		}),
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		awsv2.AWSV2Instrumentor(&conf.APIOptions)
	}

	return conf, err
}
