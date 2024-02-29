package aws

import (
	"os"
	"regexp"
	"strconv"

	"github.com/brexhq/substation/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	Region          string   `json:"region"`
	RoleARN         string   `json:"role_arn"`
	MaxRetries      int      `json:"max_retries"`
	RetryableErrors []string `json:"retryable_errors"`
}

// New returns a new AWS configuration and session.
func New(cfg Config) (*aws.Config, *session.Session) {
	conf := aws.NewConfig()

	if cfg.Region != "" {
		conf = conf.WithRegion(cfg.Region)
	} else if v, ok := os.LookupEnv("AWS_REGION"); ok {
		conf = conf.WithRegion(v)
	} else if v, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok {
		conf = conf.WithRegion(v)
	}

	retryer := NewRetryer(config.Retry{
		Count:         cfg.MaxRetries,
		ErrorMessages: cfg.RetryableErrors,
	})

	// Configurations take precedence over environment variables.
	if cfg.MaxRetries != 0 {
		goto RETRYER
	}

	if v, ok := os.LookupEnv("AWS_MAX_ATTEMPTS"); ok {
		max, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		retryer.SetMaxRetries(max)
	}

RETRYER:
	conf.Retryer = retryer
	sess := session.Must(session.NewSession())
	if cfg.RoleARN != "" {
		conf = conf.WithCredentials(stscreds.NewCredentials(sess, cfg.RoleARN))
	}

	return conf, sess
}

// NewDefault returns a new AWS configuration and session with default values.
func NewDefault() (*aws.Config, *session.Session) {
	return New(Config{})
}

func NewRetryer(cfg config.Retry) customRetryer {
	errMsg := make([]*regexp.Regexp, len(cfg.ErrorMessages))
	for i, err := range cfg.ErrorMessages {
		errMsg[i] = regexp.MustCompile(err)
	}

	return customRetryer{
		DefaultRetryer: client.DefaultRetryer{
			NumMaxRetries: cfg.Count,
		},
		errorMessages: errMsg,
	}
}

type customRetryer struct {
	client.DefaultRetryer

	// errorMessages are regular expressions that are used to match error messages.
	errorMessages []*regexp.Regexp
}

func (r customRetryer) SetMaxRetries(max int) {
	r.NumMaxRetries = max
}

// ShouldRetry retries if any of the configured error strings are found in the request error.
func (r customRetryer) ShouldRetry(req *request.Request) bool {
	for _, err := range r.errorMessages {
		if err.MatchString(req.Error.Error()) {
			return true
		}
	}

	// Fallback to the default retryer.
	return r.DefaultRetryer.ShouldRetry(req)
}
