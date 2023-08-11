package aws

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	Region     string `json:"region"`
	MaxRetries int    `json:"max_retries"`
	AssumeRole string `json:"assume_role"`
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

	if cfg.MaxRetries != 0 {
		conf = conf.WithMaxRetries(cfg.MaxRetries)
	} else if v, ok := os.LookupEnv("AWS_MAX_ATTEMPTS"); ok {
		max, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		conf = conf.WithMaxRetries(max)
	}

	sess := session.Must(session.NewSession())
	if cfg.AssumeRole != "" {
		conf = conf.WithCredentials(stscreds.NewCredentials(sess, cfg.AssumeRole))
	}

	return conf, sess
}

// NewDefault returns a new AWS configuration and session with default values.
func NewDefault() (*aws.Config, *session.Session) {
	return New(Config{})
}
