package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	ctypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	ktypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/tidwall/gjson"

	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/log"
)

const (
	// This is the period in seconds that the AWS Kinesis CloudWatch alarms
	// will evaluate the metrics over.
	kinesisMetricsPeriod = int32(60)
)

var (
	cloudwatchC *cloudwatch.Client
	kinesisC    *kinesis.Client
	// By default, AWS Kinesis streams must be below the lower threshold for
	// 100% of the evaluation period (60 minutes) to scale down. This value can
	// be overridden by the environment variable AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS.
	kinesisDownscaleDatapoints = int32(60)
	// By default, AWS Kinesis streams must be above the upper threshold for
	// 100% of the evaluation period (5 minutes) to scale up. This value can
	// be overridden by the environment variable AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS.
	kinesisUpscaleDatapoints = int32(5)
	// By default, AWS Kinesis streams will scale up if the incoming records and bytes
	// are above 70% of the threshold. This value can be overridden by the environment
	// variable AUTOSCALE_KINESIS_THRESHOLD, but it cannot be less than 40% or greater
	// than 90%.
	kinesisThreshold = 0.7
)

func init() {
	ctx := context.Background()

	awsCfg, err := iconfig.NewAWS(ctx, iconfig.AWS{})
	if err != nil {
		panic(fmt.Errorf("init: %v", err))
	}

	cloudwatchC = cloudwatch.NewFromConfig(awsCfg)
	kinesisC = kinesis.NewFromConfig(awsCfg)

	if v, found := os.LookupEnv("AUTOSCALE_KINESIS_DOWNSCALE_DATAPOINTS"); found {
		dps, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		kinesisDownscaleDatapoints = int32(dps)
	}

	if v, found := os.LookupEnv("AUTOSCALE_KINESIS_UPSCALE_DATAPOINTS"); found {
		dps, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		kinesisUpscaleDatapoints = int32(dps)
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

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, snsEvent events.SNSEvent) error {
	payload := snsEvent.Records[0].SNS
	topicArn := payload.TopicArn
	message := payload.Message

	alarmName := gjson.Get(message, "AlarmName").String()
	triggerMetrics := gjson.Get(message, "Trigger.Metrics")

	log.WithField("alarm", alarmName).Debug("Received autoscale notification.")

	var stream string
	for _, v := range triggerMetrics.Array() {
		id := gjson.Get(v.String(), "Id").String()
		if id == "m1" || id == "m2" {
			stream = gjson.Get(v.String(), "MetricStat.Metric.Dimensions.0.value").String()
			break
		}
	}
	log.WithField("alarm", alarmName).WithField("stream", stream).Debug("Parsed Kinesis stream.")

	shards, err := listShards(ctx, stream)
	if err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", shards).
		Info("Retrieved active shard count.")

	var newShards int32
	if strings.Contains(alarmName, "upscale") {
		newShards = upscale(float64(shards))
	}
	if strings.Contains(alarmName, "downscale") {
		newShards = downscale(float64(shards))
	}

	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", newShards).Info("Calculated new shard count.")

	tags, err := listTags(ctx, stream)
	if err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	var minShard, maxShard int32
	for _, tag := range tags {
		if *tag.Key == "MinimumShards" {
			minShard, err := strconv.ParseInt(*tag.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("handler: %v", err)
			}

			log.WithField("stream", stream).WithField("count", minShard).Debug("Retrieved minimum shard count.")
		}

		if *tag.Key == "MaximumShards" {
			maxShard, err := strconv.ParseInt(*tag.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("handler: %v", err)
			}

			log.WithField("stream", stream).WithField("count", maxShard).Debug("Retrieved maximum shard count.")
		}

		// Tracking the last scaling event prevents scaling from occurring too frequently.
		// If the current scaling event is an upscale, then the last scaling event must be at least 3 minutes ago.
		// If the current scaling event is a downscale, then the last scaling event must be at least 30 minutes ago.
		if *tag.Key == "LastScalingEvent" {
			lastScalingEvent, err := time.Parse(time.RFC3339, *tag.Value)
			if err != nil {
				return fmt.Errorf("handler: %v", err)
			}

			if (time.Since(lastScalingEvent) < 3*time.Minute && strings.Contains(alarmName, "upscale")) ||
				(time.Since(lastScalingEvent) < 30*time.Minute && strings.Contains(alarmName, "downscale")) {
				log.WithField("stream", stream).WithField("time", lastScalingEvent).Info("Last scaling event is too recent.")

				if _, err := cloudwatchC.SetAlarmState(ctx, &cloudwatch.SetAlarmStateInput{
					AlarmName:   aws.String(alarmName),
					StateValue:  ctypes.StateValueInsufficientData,
					StateReason: aws.String("Last scaling event is too recent"),
				}); err != nil {
					return fmt.Errorf("handler: %v", err)
				}

				return nil
			}
		}
	}

	if minShard != 0 && newShards < minShard {
		newShards = minShard
	}

	if maxShard != 0 && newShards > maxShard {
		newShards = maxShard
	}

	if newShards < 1 {
		newShards = 1
	}

	if newShards == shards {
		log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", shards).Info("Active shard count is at minimum threshold, no change is required.")
		return nil
	}

	if err := updateStream(ctx, stream, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", newShards).Info("Updated shard count.")

	metrics := []ctypes.MetricDataQuery{
		{
			Id: aws.String("m1"),
			MetricStat: &ctypes.MetricStat{
				Metric: &ctypes.Metric{
					Namespace:  aws.String("AWS/Kinesis"),
					MetricName: aws.String("IncomingRecords"),
					Dimensions: []ctypes.Dimension{
						{
							Name:  aws.String("StreamName"),
							Value: aws.String(stream),
						},
					},
				},
				Period: aws.Int32(kinesisMetricsPeriod),
				Stat:   aws.String("Sum"),
			},
			Label:      aws.String("IncomingRecords"),
			ReturnData: aws.Bool(false),
		},
		{
			Id: aws.String("m2"),
			MetricStat: &ctypes.MetricStat{
				Metric: &ctypes.Metric{
					Namespace:  aws.String("AWS/Kinesis"),
					MetricName: aws.String("IncomingBytes"),
					Dimensions: []ctypes.Dimension{
						{
							Name:  aws.String("StreamName"),
							Value: aws.String(stream),
						},
					},
				},
				Period: aws.Int32(kinesisMetricsPeriod),
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
				fmt.Sprintf("e1/(1000*%d*%d)", newShards, kinesisMetricsPeriod),
			),
			Label:      aws.String("IncomingRecordsPercent"),
			ReturnData: aws.Bool(false),
		},
		{
			Id: aws.String("e4"),
			Expression: aws.String(
				fmt.Sprintf("e2/(1048576*%d*%d)", newShards, kinesisMetricsPeriod),
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
	}

	downscaleThreshold := kinesisThreshold - 0.35
	if err := updateDownscaleAlarm(ctx, stream, topicArn, downscaleThreshold, metrics); err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	log.WithField("alarm", stream+"_downscale").WithField("stream", stream).WithField("count", newShards).Debug("Reset CloudWatch alarm.")

	upscaleThreshold := kinesisThreshold
	if err := updateUpscaleAlarm(ctx, stream, topicArn, upscaleThreshold, metrics); err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	log.WithField("alarm", stream+"_upscale").WithField("stream", stream).WithField("count", newShards).Debug("Reset CloudWatch alarm.")

	return nil
}

func downscale(shards float64) int32 {
	switch {
	case shards < 5:
		return int32(math.Ceil(shards / 2))
	case shards < 13:
		return int32(math.Ceil(shards / 1.75))
	case shards < 33:
		return int32(math.Ceil(shards / 1.5))
	default:
		return int32(math.Ceil(shards / 1.25))
	}
}

func upscale(shards float64) int32 {
	switch {
	case shards < 5:
		return int32(math.Floor(shards * 2))
	case shards < 13:
		return int32(math.Floor(shards * 1.75))
	case shards < 33:
		return int32(math.Floor(shards * 1.5))
	default:
		return int32(math.Floor(shards * 1.25))
	}
}

func listShards(ctx context.Context, stream string) (int32, error) {
	var shards int32

	input := kinesis.ListShardsInput{
		StreamName: aws.String(stream),
	}

LOOP:
	for {
		resp, err := kinesisC.ListShards(ctx, &input)
		if err != nil {
			return 0, err
		}

		for _, s := range resp.Shards {
			if end := s.SequenceNumberRange.EndingSequenceNumber; end == nil {
				shards++
			}
		}

		if resp.NextToken != nil {
			input = kinesis.ListShardsInput{
				NextToken: resp.NextToken,
			}
		} else {
			break LOOP
		}
	}

	return shards, nil
}

func listTags(ctx context.Context, stream string) ([]ktypes.Tag, error) {
	var tags []ktypes.Tag
	var lastTag string

	for {
		input := kinesis.ListTagsForStreamInput{
			StreamName: aws.String(stream),
		}

		if lastTag != "" {
			input.ExclusiveStartTagKey = aws.String(lastTag)
		}

		resp, err := kinesisC.ListTagsForStream(ctx, &input)
		if err != nil {
			return nil, err
		}

		if len(resp.Tags) == 0 {
			break
		}

		tags = append(tags, resp.Tags...)
		lastTag = *resp.Tags[len(resp.Tags)-1].Key

		if !*resp.HasMoreTags {
			break
		}
	}

	return tags, nil
}

func updateStream(ctx context.Context, stream string, shards int32) error {
	_, err := kinesisC.UpdateShardCount(ctx, &kinesis.UpdateShardCountInput{
		StreamName:       aws.String(stream),
		TargetShardCount: aws.Int32(shards),
		ScalingType:      ktypes.ScalingTypeUniformScaling,
	})
	if err != nil {
		return err
	}

	for {
		resp, err := kinesisC.DescribeStreamSummary(ctx, &kinesis.DescribeStreamSummaryInput{
			StreamName: aws.String(stream),
		})
		if err != nil {
			return err
		}

		if resp.StreamDescriptionSummary.StreamStatus != ktypes.StreamStatusUpdating {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if _, err := kinesisC.AddTagsToStream(ctx, &kinesis.AddTagsToStreamInput{
		StreamName: aws.String(stream),
		Tags: map[string]string{
			"LastScalingEvent": time.Now().Format(time.RFC3339),
		},
	}); err != nil {
		return err
	}

	return nil
}

func updateDownscaleAlarm(ctx context.Context, stream, topic string, threshold float64, metrics []ctypes.MetricDataQuery) error {
	if _, err := cloudwatchC.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(stream + "_downscale"),
		AlarmDescription:   aws.String(stream),
		ActionsEnabled:     aws.Bool(true),
		AlarmActions:       []string{topic},
		EvaluationPeriods:  aws.Int32(kinesisDownscaleDatapoints),
		DatapointsToAlarm:  aws.Int32(kinesisDownscaleDatapoints),
		Threshold:          aws.Float64(threshold),
		ComparisonOperator: ctypes.ComparisonOperatorLessThanOrEqualToThreshold,
		TreatMissingData:   aws.String("ignore"),
		Metrics:            metrics,
	}); err != nil {
		return err
	}

	if _, err := cloudwatchC.SetAlarmState(ctx, &cloudwatch.SetAlarmStateInput{
		AlarmName:   aws.String(stream + "_downscale"),
		StateValue:  ctypes.StateValueInsufficientData,
		StateReason: aws.String("Threshold updated"),
	}); err != nil {
		return err
	}

	return nil
}

func updateUpscaleAlarm(ctx context.Context, stream, topic string, threshold float64, metrics []ctypes.MetricDataQuery) error {
	if _, err := cloudwatchC.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(stream + "_upscale"),
		AlarmDescription:   aws.String(stream),
		ActionsEnabled:     aws.Bool(true),
		AlarmActions:       []string{topic},
		EvaluationPeriods:  aws.Int32(kinesisUpscaleDatapoints),
		DatapointsToAlarm:  aws.Int32(kinesisUpscaleDatapoints),
		Threshold:          aws.Float64(threshold),
		ComparisonOperator: ctypes.ComparisonOperatorGreaterThanOrEqualToThreshold,
		TreatMissingData:   aws.String("ignore"),
		Metrics:            metrics,
	}); err != nil {
		return err
	}

	if _, err := cloudwatchC.SetAlarmState(ctx, &cloudwatch.SetAlarmStateInput{
		AlarmName:   aws.String(stream + "_upscale"),
		StateValue:  ctypes.StateValueInsufficientData,
		StateReason: aws.String("Threshold updated"),
	}); err != nil {
		return err
	}

	return nil
}
