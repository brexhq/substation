// package config provides configuration types and functions for Substation.
//
// Any non-backwards compatible changes to the configuration types should be
// accompanied by a version bump. Use the guidance below for choosing the
// appropriate fields for configurations:
//
// For time-based configurations:
//
//   - Use `Delay` for the amount of time to wait before executing.
//
//   - Use `Timeout` for the amount of time to wait before interrupting
//     an execution.
//
//   - Use `Duration` for the total amount of time over many executions.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"

	"github.com/brexhq/substation/v2/config"
)

var (
	// ErrInvalidFactoryInput is returned when an unsupported input is referenced in any factory function.
	ErrInvalidFactoryInput = fmt.Errorf("invalid factory input")

	// ErrMissingRequiredOption is returned when a component does not have the required options to properly run.
	ErrMissingRequiredOption = fmt.Errorf("missing required option")

	// ErrInvalidOption is returned when an invalid option is received in a constructor.
	ErrInvalidOption = fmt.Errorf("invalid option")
)

type Object struct {
	// SourceKey retrieves a value from a JSON object.
	SourceKey string `json:"source_key"`
	// TargetKey place a value into a JSON object.
	TargetKey string `json:"target_key"`
	// BatchKey retrieves a value from a JSON object that is used to organize
	// batched data (internal/aggregate).
	BatchKey string `json:"batch_key"`
}

type Metric struct {
	// Name is the name of the metric.
	Name string `json:"name"`
	// Attributes are key-value pairs that are associated with the metric.
	Attributes map[string]string `json:"attributes"`
	// Destination is the metrics destination that the metric will be sent to (internal/metrics).
	Destination config.Config `json:"destination"`
}

type Request struct {
	// Timeout is the amount of time that the request will wait before timing out.
	Timeout string `json:"Timeout"`
}

type Retry struct {
	// Count is the maximum number of times that the action will be retried. This
	// can be combined with the Delay field to create a backoff strategy.
	Count int `json:"count"`
	// Delay is the amount of time to wait before retrying the action. This can be
	// combined with the Count field to create a backoff strategy.
	Delay string `json:"delay"`
}

type Batch struct {
	// Count is the maximum number of records that can be batched.
	Count int `json:"count"`
	// Size is the maximum size of the batch in bytes.
	Size int `json:"size"`
	// Duration is the maximum amount of time that records can be batched for.
	Duration string `json:"duration"`
}

// Decode marshals and unmarshals an input interface into the output interface
// using the standard library's json package. This should be used when decoding
// JSON configurations (i.e., Config) in Substation interface factories.
func Decode(input, output interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, output)
}

type AWS struct {
	// ARN is the AWS resource that the action will interact with.
	ARN string `json:"arn"`
	// AssumeRoleARN is the ARN of the role that the action will assume.
	AssumeRoleARN string `json:"assume_role_arn"`
}

// NewAWS returns a valid AWS SDK v2 configuration.
func NewAWS(ctx context.Context, cfg AWS) (aws.Config, error) {
	arnx, _ := arn.Parse(cfg.ARN)           // Ignore missing ARN errors.
	arny, _ := arn.Parse(cfg.AssumeRoleARN) // Ignore missing ARN errors.

	var region string
	if arnx.Region != "" {
		region = arnx.Region
	} else if arny.Region != "" {
		region = arny.Region
	} else if v, ok := os.LookupEnv("AWS_REGION"); ok {
		region = v
	} else if v, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok {
		region = v
	}

	var creds aws.CredentialsProvider // nil is a valid default.
	if cfg.AssumeRoleARN != "" {
		conf, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
		)
		if err != nil {
			return aws.Config{}, err
		}

		stsSvc := sts.NewFromConfig(conf)
		creds = stscreds.NewAssumeRoleProvider(stsSvc, cfg.AssumeRoleARN)
	}

	maxRetry := 3 // Matches the standard retryer.
	if v, ok := os.LookupEnv("AWS_MAX_ATTEMPTS"); ok {
		max, err := strconv.Atoi(v)
		if err != nil {
			return aws.Config{}, err
		}

		maxRetry = max
	}

	conf, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.MaxAttempts = maxRetry
			})
		}),
	)

	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		awsv2.AWSV2Instrumentor(&conf.APIOptions)
	}

	return conf, err
}
