package cloudwatch

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	_aws "github.com/brexhq/substation/internal/aws"
)

const (
	kinesisMetricsPeriod = 60
	// AWS Kinesis streams will scale down / in if they are less than 25% of the Kinesis service limits within a 60 minute / 1 hour period.
	kinesisDownscaleEvaluationPeriod, kinesisDownscaleThreshold = 60, 0.25
	// AWS Kinesis streams will scale up / out if they are greater than 75% of the Kinesis service limits within a 5 minute period.
	kinesisUpscaleEvaluationPeriod, kinesisUpscaleThreshold = 5, 0.75
)

var (
	// By default, AWS Kinesis streams must be below the lower threshold for 95% of the evaluation period (57 minutes) to scale down. This value can be overridden by the environment variable SUBSTATION_AUTOSCALING_DOWNSCALE_DATAPOINTS, but it cannot exceed 60 minutes.
	kinesisDownscaleDatapoints = 57
	// By default, AWS Kinesis streams must be above the upper threshold for 100% of the evaluation period (5 minutes) to scale up. This value can be overridden by the environment variable SUBSTATION_AUTOSCALING_UPSCALE_DATAPOINTS, but it cannot exceed 5 minutes.
	kinesisUpscaleDatapoints = 5
)

func init() {
	if v, found := os.LookupEnv("SUBSTATION_AUTOSCALING_DOWNSCALE_DATAPOINTS"); found {
		downscale, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		if downscale <= kinesisDownscaleEvaluationPeriod {
			kinesisDownscaleDatapoints = downscale
		}
	}

	if v, found := os.LookupEnv("SUBSTATION_AUTOSCALING_UPSCALE_DATAPOINTS"); found {
		upscale, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		if upscale <= kinesisUpscaleEvaluationPeriod {
			kinesisUpscaleDatapoints = upscale
		}
	}
}

// New returns a configured CloudWatch client.
func New(cfg _aws.Config) *cloudwatch.CloudWatch {
	conf, sess := _aws.New(cfg)

	c := cloudwatch.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the CloudWatch API interface.
type API struct {
	Client cloudwatchiface.CloudWatchAPI
}

// Setup creates a new CloudWatch client.
func (a *API) Setup(cfg _aws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// UpdateKinesisDownscaleAlarm updates CloudWatch alarms that manage the scale down tracking for Kinesis streams.
func (a *API) UpdateKinesisDownscaleAlarm(ctx aws.Context, name, stream, topic string, shards int64) error {
	if _, err := a.Client.PutMetricAlarmWithContext(
		ctx,
		&cloudwatch.PutMetricAlarmInput{
			AlarmName:          aws.String(name),
			AlarmDescription:   aws.String(stream),
			ActionsEnabled:     aws.Bool(true),
			AlarmActions:       []*string{aws.String(topic)},
			EvaluationPeriods:  aws.Int64(kinesisDownscaleEvaluationPeriod),
			DatapointsToAlarm:  aws.Int64(int64(kinesisDownscaleDatapoints)),
			Threshold:          aws.Float64(kinesisDownscaleThreshold),
			ComparisonOperator: aws.String("LessThanOrEqualToThreshold"),
			TreatMissingData:   aws.String("ignore"),
			Metrics: []*cloudwatch.MetricDataQuery{
				{
					Id: aws.String("m1"),
					MetricStat: &cloudwatch.MetricStat{
						Metric: &cloudwatch.Metric{
							Namespace:  aws.String("AWS/Kinesis"),
							MetricName: aws.String("IncomingRecords"),
							Dimensions: []*cloudwatch.Dimension{
								{
									Name:  aws.String("StreamName"),
									Value: aws.String(stream),
								},
							},
						},
						Period: aws.Int64(kinesisMetricsPeriod),
						Stat:   aws.String("Sum"),
					},
					Label:      aws.String("IncomingRecords"),
					ReturnData: aws.Bool(false),
				},
				{
					Id: aws.String("m2"),
					MetricStat: &cloudwatch.MetricStat{
						Metric: &cloudwatch.Metric{
							Namespace:  aws.String("AWS/Kinesis"),
							MetricName: aws.String("IncomingBytes"),
							Dimensions: []*cloudwatch.Dimension{
								{
									Name:  aws.String("StreamName"),
									Value: aws.String(stream),
								},
							},
						},
						Period: aws.Int64(kinesisMetricsPeriod),
						Stat:   aws.String("Sum"),
					},
					Label:      aws.String("IncomingBytes"),
					ReturnData: aws.Bool(false),
				},
				{
					Id:         aws.String("e1"),
					Expression: aws.String("FILL(m1,REPEAT)"),
					Label:      aws.String("FillMissingDataPointsForIncomingRecords"),
					ReturnData: aws.Bool(false),
				},
				{
					Id:         aws.String("e2"),
					Expression: aws.String("FILL(m2,REPEAT)"),
					Label:      aws.String("FillMissingDataPointsForIncomingBytes"),
					ReturnData: aws.Bool(false),
				},
				{
					Id: aws.String("e3"),
					Expression: aws.String(
						fmt.Sprintf("e1/(1000*%d*%d)", shards, kinesisMetricsPeriod),
					),
					Label:      aws.String("IncomingRecordsPercent"),
					ReturnData: aws.Bool(false),
				},
				{
					Id: aws.String("e4"),
					Expression: aws.String(
						fmt.Sprintf("e2/(1048576*%d*%d)", shards, kinesisMetricsPeriod),
					),
					Label:      aws.String("IncomingBytesPercent"),
					ReturnData: aws.Bool(false),
				},
				{
					Id:         aws.String("e5"),
					Expression: aws.String("MAX([e3,e4])"),
					Label:      aws.String("IncomingMax"),
					ReturnData: aws.Bool(true),
				},
			},
		}); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %v", name, stream, err)
	}

	if _, err := a.Client.SetAlarmStateWithContext(ctx,
		&cloudwatch.SetAlarmStateInput{
			AlarmName:   aws.String(name),
			StateValue:  aws.String("INSUFFICIENT_DATA"),
			StateReason: aws.String("Threshold value updated"),
		}); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %v", name, stream, err)
	}

	return nil
}

// UpdateKinesisUpscaleAlarm updates CloudWatch alarms that manage the scale up tracking for Kinesis streams.
func (a *API) UpdateKinesisUpscaleAlarm(ctx aws.Context, name, stream, topic string, shards int64) error {
	if _, err := a.Client.PutMetricAlarmWithContext(
		ctx,
		&cloudwatch.PutMetricAlarmInput{
			AlarmName:          aws.String(name),
			AlarmDescription:   aws.String(stream),
			ActionsEnabled:     aws.Bool(true),
			AlarmActions:       []*string{aws.String(topic)},
			EvaluationPeriods:  aws.Int64(kinesisUpscaleEvaluationPeriod),
			DatapointsToAlarm:  aws.Int64(int64(kinesisUpscaleDatapoints)),
			Threshold:          aws.Float64(kinesisUpscaleThreshold),
			ComparisonOperator: aws.String("GreaterThanOrEqualToThreshold"),
			TreatMissingData:   aws.String("ignore"),
			Metrics: []*cloudwatch.MetricDataQuery{
				{
					Id: aws.String("m1"),
					MetricStat: &cloudwatch.MetricStat{
						Metric: &cloudwatch.Metric{
							Namespace:  aws.String("AWS/Kinesis"),
							MetricName: aws.String("IncomingRecords"),
							Dimensions: []*cloudwatch.Dimension{
								{
									Name:  aws.String("StreamName"),
									Value: aws.String(stream),
								},
							},
						},
						Period: aws.Int64(kinesisMetricsPeriod),
						Stat:   aws.String("Sum"),
					},
					Label:      aws.String("IncomingRecords"),
					ReturnData: aws.Bool(false),
				},
				{
					Id: aws.String("m2"),
					MetricStat: &cloudwatch.MetricStat{
						Metric: &cloudwatch.Metric{
							Namespace:  aws.String("AWS/Kinesis"),
							MetricName: aws.String("IncomingBytes"),
							Dimensions: []*cloudwatch.Dimension{
								{
									Name:  aws.String("StreamName"),
									Value: aws.String(stream),
								},
							},
						},
						Period: aws.Int64(kinesisMetricsPeriod),
						Stat:   aws.String("Sum"),
					},
					Label:      aws.String("IncomingBytes"),
					ReturnData: aws.Bool(false),
				},
				{
					Id:         aws.String("e1"),
					Expression: aws.String("FILL(m1,REPEAT)"),
					Label:      aws.String("FillMissingDataPointsForIncomingRecords"),
					ReturnData: aws.Bool(false),
				},
				{
					Id:         aws.String("e2"),
					Expression: aws.String("FILL(m2,REPEAT)"),
					Label:      aws.String("FillMissingDataPointsForIncomingBytes"),
					ReturnData: aws.Bool(false),
				},
				{
					Id: aws.String("e3"),
					Expression: aws.String(
						fmt.Sprintf("e1/(1000*%d*%d)", shards, kinesisMetricsPeriod),
					),
					Label:      aws.String("IncomingRecordsPercent"),
					ReturnData: aws.Bool(false),
				},
				{
					Id: aws.String("e4"),
					Expression: aws.String(
						fmt.Sprintf("e2/(1048576*%d*%d)", shards, kinesisMetricsPeriod),
					),
					Label:      aws.String("IncomingBytesPercent"),
					ReturnData: aws.Bool(false),
				},
				{
					Id:         aws.String("e5"),
					Expression: aws.String("MAX([e3,e4])"),
					Label:      aws.String("IncomingMax"),
					ReturnData: aws.Bool(true),
				},
			},
		}); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %v", name, stream, err)
	}

	if _, err := a.Client.SetAlarmStateWithContext(ctx,
		&cloudwatch.SetAlarmStateInput{
			AlarmName:   aws.String(name),
			StateValue:  aws.String("INSUFFICIENT_DATA"),
			StateReason: aws.String("Threshold value updated"),
		}); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %v", name, stream, err)
	}

	return nil
}
