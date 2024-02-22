package aws

import (
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	Region     string `json:"region"`
	MaxRetries int    `json:"max_retries"`
	RoleARN    string `json:"role_arn"`
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
		conf.Retryer = CustomRetryer{
			DefaultRetryer: client.DefaultRetryer{
				NumMaxRetries: cfg.MaxRetries,
			},
		}
	} else if v, ok := os.LookupEnv("AWS_MAX_ATTEMPTS"); ok {
		max, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		conf.Retryer = CustomRetryer{
			DefaultRetryer: client.DefaultRetryer{
				NumMaxRetries: max,
			},
		}
	}

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

type CustomRetryer struct {
	client.DefaultRetryer
}

// ShouldRetry will retry if the connection is reset by the peer.
func (r CustomRetryer) ShouldRetry(req *request.Request) bool {
	if strings.Contains(req.Error.Error(), "read: connection reset by peer") {
		return true
	}

	// Fallback to SDK's built in retry rules
	return r.DefaultRetryer.ShouldRetry(req)
}
