package cloudwatch

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	iaws "github.com/brexhq/substation/internal/aws"
)

const (
	// This is the period in seconds that the AWS Kinesis CloudWatch alarms
	// will evaluate the metrics over.
	kinesisMetricsPeriod = 60
)

var (
	// By default, AWS Kinesis streams must be below the lower threshold for
	// 100% of the evaluation period (60 minutes) to scale down. This value can
	// be overridden by the environment variable AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS.
	// If overridden, then the evaluation period will be adjusted every 60 minutes,
	// up to 6 hours, as needed to match the number of datapoints.
	kinesisDownscaleDatapoints = 60
	// By default, AWS Kinesis streams must be above the upper threshold for
	// 100% of the evaluation period (5 minutes) to scale up. This value can
	// be overridden by the environment variable AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS.
	// If overridden, then the evaluation period will be adjusted every 5 minutes,
	// up to 30 minutes, as needed to match the number of datapoints.
	kinesisUpscaleDatapoints = 5
	// By default, AWS Kinesis streams will scale up if the incoming records and bytes
	// are above 70% of the threshold. This value can be overridden by the environment
	// variable AUTOSCALE_KINESIS_THRESHOLD, but it cannot be less than 40% or greater
	// than 90%.
	kinesisThreshold = 0.7
)

func init() {
	if v, found := os.LookupEnv("AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS"); found {
		dps, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		kinesisDownscaleDatapoints = dps
	}

	if v, found := os.LookupEnv("AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS"); found {
		dps, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		kinesisUpscaleDatapoints = dps
	}

	if v, found := os.LookupEnv("AUTOSCALE_KINESIS_THRESHOLD"); found {
		threshold, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(err)
		}

		if threshold >= 0.4 && threshold <= 0.9 {
			kinesisThreshold = threshold
		}
	}
}

// New returns a configured CloudWatch client.
func New(cfg iaws.Config) *cloudwatch.CloudWatch {
	conf, sess := iaws.New(cfg)

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
func (a *API) Setup(cfg iaws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// UpdateKinesisDownscaleAlarm updates CloudWatch alarms that manage the scale down tracking for Kinesis streams.
func (a *API) UpdateKinesisDownscaleAlarm(ctx aws.Context, name, stream, topic string, shards int64) error {
	downscaleThreshold := kinesisThreshold - 0.35

	if _, err := a.Client.PutMetricAlarmWithContext(
		ctx,
		&cloudwatch.PutMetricAlarmInput{
			AlarmName:          aws.String(name),
			AlarmDescription:   aws.String(stream),
			ActionsEnabled:     aws.Bool(true),
			AlarmActions:       []*string{aws.String(topic)},
			EvaluationPeriods:  aws.Int64(int64(kinesisDownscaleDatapoints)),
			DatapointsToAlarm:  aws.Int64(int64(kinesisDownscaleDatapoints)),
			Threshold:          aws.Float64(downscaleThreshold),
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

	if err := a.UpdateKinesisAlarmState(ctx, name, "Threshold value updated"); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %v", name, stream, err)
	}

	return nil
}

// UpdateKinesisUpscaleAlarm updates CloudWatch alarms that manage the scale up tracking for Kinesis streams.
func (a *API) UpdateKinesisUpscaleAlarm(ctx aws.Context, name, stream, topic string, shards int64) error {
	upscaleThreshold := kinesisThreshold

	if _, err := a.Client.PutMetricAlarmWithContext(
		ctx,
		&cloudwatch.PutMetricAlarmInput{
			AlarmName:          aws.String(name),
			AlarmDescription:   aws.String(stream),
			ActionsEnabled:     aws.Bool(true),
			AlarmActions:       []*string{aws.String(topic)},
			EvaluationPeriods:  aws.Int64(int64(kinesisUpscaleDatapoints)),
			DatapointsToAlarm:  aws.Int64(int64(kinesisUpscaleDatapoints)),
			Threshold:          aws.Float64(upscaleThreshold),
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

	if err := a.UpdateKinesisAlarmState(ctx, name, "Threshold value updated"); err != nil {
		return fmt.Errorf("updatealarm alarm %s stream %s: %v", name, stream, err)
	}

	return nil
}

func (a *API) UpdateKinesisAlarmState(ctx aws.Context, name, reason string) error {
	_, err := a.Client.SetAlarmStateWithContext(ctx,
		&cloudwatch.SetAlarmStateInput{
			AlarmName:   aws.String(name),
			StateValue:  aws.String("INSUFFICIENT_DATA"),
			StateReason: aws.String(reason),
		})
	return err
}
