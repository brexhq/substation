package cloudwatch

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

const (
	kinesisMetricsPeriod = 60
	// scale Kinesis stream down if it is below threshold for 1 hour
	kinesisDownscaleEvaluationPeriod = 1 * 60
	kinesisDownscaleThreshold        = 0.25
	// scale Kinesis stream up if it is above threshold for 5 minutes
	kinesisUpscaleEvaluationPeriod = 1 * 5
	kinesisUpscaleThreshold        = 0.75
)

// New returns a configured CloudWatch client.
func New() *cloudwatch.CloudWatch {
	conf := aws.NewConfig()

	// provides forward compatibility for the Go SDK to support env var configuration settings
	// https://github.com/aws/aws-sdk-go/issues/4207
	max, found := os.LookupEnv("AWS_MAX_ATTEMPTS")
	if found {
		m, err := strconv.Atoi(max)
		if err != nil {
			panic(err)
		}

		conf = conf.WithMaxRetries(m)
	}

	c := cloudwatch.New(
		session.Must(session.NewSession()),
		conf,
	)

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
func (a *API) Setup() {
	a.Client = New()
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
			DatapointsToAlarm:  aws.Int64(kinesisDownscaleEvaluationPeriod * 0.95),
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
		return fmt.Errorf("updatealarm alarm %s stream %s: %w", name, stream, err)
	}

	if _, err := a.Client.SetAlarmStateWithContext(ctx,
		&cloudwatch.SetAlarmStateInput{
			AlarmName:   aws.String(name),
			StateValue:  aws.String("INSUFFICIENT_DATA"),
			StateReason: aws.String("Threshold value updated"),
		}); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %w", name, stream, err)
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
			DatapointsToAlarm:  aws.Int64(kinesisUpscaleEvaluationPeriod),
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
		return fmt.Errorf("updatealarm alarm %s stream %s: %w", name, stream, err)
	}

	if _, err := a.Client.SetAlarmStateWithContext(ctx,
		&cloudwatch.SetAlarmStateInput{
			AlarmName:   aws.String(name),
			StateValue:  aws.String("INSUFFICIENT_DATA"),
			StateReason: aws.String("Threshold value updated"),
		}); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %w", name, stream, err)
	}

	return nil
}
