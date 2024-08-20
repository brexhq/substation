package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/v2/config"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/tidwall/sjson"
)

type awsCloudWatchEmbeddedMetricsConfig struct{}

// awsCloudWatchEmbeddedMetrics creates a metric in the AWS Embedded Metrics
// Format and writes it to standard output. This is the preferred method for
// generating metrics from AWS Lambda functions. Read more about the Embedded
// Metrics Format specification here:
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html.
type awsCloudWatchEmbeddedMetrics struct {
	conf awsCloudWatchEmbeddedMetricsConfig
}

func newAWSCloudWatchEmbeddedMetrics(_ context.Context, cfg config.Config) (*awsCloudWatchEmbeddedMetrics, error) {
	conf := awsCloudWatchEmbeddedMetricsConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	return &awsCloudWatchEmbeddedMetrics{
		conf: conf,
	}, nil
}

func (m *awsCloudWatchEmbeddedMetrics) Generate(ctx context.Context, data Data) (err error) {
	emf := []byte{}

	emf, err = sjson.SetBytes(emf, "_aws.Timestamp", time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	emf, err = sjson.SetBytes(emf, "_aws.CloudWatchMetrics.0.Namespace", metricsApplication)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	var dimensions []string
	for key, val := range data.Attributes {
		dimensions = append(dimensions, key)

		emf, err = sjson.SetBytes(emf, key, val)
		if err != nil {
			return fmt.Errorf("metrics log_embedded_metrics: %v", err)
		}
	}

	emf, err = sjson.SetBytes(emf, "_aws.CloudWatchMetrics.0.Dimensions.-1", dimensions)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	emf, err = sjson.SetBytes(emf, "_aws.CloudWatchMetrics.0.Metrics.0.Name", data.Name)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	emf, err = sjson.SetBytes(emf, data.Name, data.Value)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	// Logging EMF to standard out in AWS Lambda automatically sends metrics to CloudWatch.
	fmt.Println(string(emf))

	return nil
}
