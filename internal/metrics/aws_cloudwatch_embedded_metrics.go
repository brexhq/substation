package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/brexhq/substation/internal/json"
)

// AWSCloudWatchEmbeddedMetrics creates a metric in the AWS Embedded Metrics Format and writes it to standard output. This is the preferred method for generating metrics from AWS Lambda functions. Read more about the Embedded Metrics Format specification here: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html.
type AWSCloudWatchEmbeddedMetrics struct{}

/*
Generate creates a metric with the AWSCloudWatchEmbeddedMetrics metrics generator. All Attributes in the metrics.Data struct are inserted as CloudWatch Metrics dimensions; if the generator is invoked from an AWS Lambda function, then the function name is automatically added as a dimension. This method creates a JSON object with the structure shown below, where references are filled in from the metrics.Data struct:

	{
		"_aws": {
			"Timestamp": $currentTime,
			"CloudWatchMetrics": [
				{
					"Namespace": $metricsApplication,
					"Dimensions": [
						[
							$data.Attributes.key
						]
					],
					"Name": $data.Name,
				}
			]
		},
		$data.Attributes.key: $data.Attributes.value,
		$data.Name: $data.Value
	}
*/
func (m AWSCloudWatchEmbeddedMetrics) Generate(ctx context.Context, data Data) (err error) {
	emf := []byte{}

	// default values for CloudWatch metrics from Substation applications
	// if the metrics are generated from AWS Lambda, then the function name is automatically tagged
	emf, err = json.Set(emf, "_aws.Timestamp", time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	emf, err = json.Set(emf, "_aws.CloudWatchMetrics.0.Namespace", metricsApplication)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	if metricsAWSLambdaFunctionName != "" {
		attr := map[string]string{"FunctionName": metricsAWSLambdaFunctionName}
		data.AddAttributes(attr)
	}

	for key, val := range data.Attributes {
		emf, err = json.Set(emf, "_aws.CloudWatchMetrics.0.Dimensions.-1.-1", key)
		if err != nil {
			return fmt.Errorf("metrics log_embedded_metrics: %v", err)
		}

		emf, err = json.Set(emf, key, val)
		if err != nil {
			return fmt.Errorf("metrics log_embedded_metrics: %v", err)
		}
	}

	emf, err = json.Set(emf, "_aws.CloudWatchMetrics.0.Metrics.0.Name", data.Name)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	emf, err = json.Set(emf, data.Name, data.Value)
	if err != nil {
		return fmt.Errorf("metrics log_embedded_metrics: %v", err)
	}

	// logging EMF to standard out in AWS Lambda automatically sends metrics to CloudWatch
	fmt.Println(string(emf))

	return nil
}
