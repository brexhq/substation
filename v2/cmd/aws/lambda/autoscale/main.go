package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brexhq/substation/v2/internal/aws"
	"github.com/brexhq/substation/v2/internal/aws/cloudwatch"
	"github.com/brexhq/substation/v2/internal/aws/kinesis"
	"github.com/brexhq/substation/v2/internal/log"
	"github.com/tidwall/gjson"
)

var (
	cloudwatchAPI cloudwatch.API
	kinesisAPI    kinesis.API
)

func init() {
	// These must run in the same AWS account and region as the Lambda function.
	cloudwatchAPI.Setup(aws.Config{})
	kinesisAPI.Setup(aws.Config{})
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

	shards, err := kinesisAPI.ActiveShards(ctx, stream)
	if err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", shards).
		Info("Retrieved active shard count.")

	var newShards int64
	if strings.Contains(alarmName, "upscale") {
		newShards = upscale(float64(shards))
	}
	if strings.Contains(alarmName, "downscale") {
		newShards = downscale(float64(shards))
	}

	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", newShards).Info("Calculated new shard count.")

	tags, err := kinesisAPI.GetTags(ctx, stream)
	if err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	var minShard, maxShard int64
	for _, tag := range tags {
		if *tag.Key == "MinimumShards" {
			minShard, err = strconv.ParseInt(*tag.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("handler: %v", err)
			}

			log.WithField("stream", stream).WithField("count", minShard).Debug("Retrieved minimum shard count.")
		}

		if *tag.Key == "MaximumShards" {
			maxShard, err = strconv.ParseInt(*tag.Value, 10, 64)
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

				if err := cloudwatchAPI.UpdateKinesisAlarmState(ctx, alarmName, "Last scaling event is too recent"); err != nil {
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

	if err := kinesisAPI.UpdateShards(ctx, stream, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	if err := kinesisAPI.UpdateTag(ctx, stream, "LastScalingEvent", time.Now().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("handler: %v", err)
	}

	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", newShards).Info("Updated shard count.")

	if err := cloudwatchAPI.UpdateKinesisDownscaleAlarm(ctx, stream+"_downscale", stream, topicArn, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", stream+"_downscale").WithField("stream", stream).WithField("count", newShards).Debug("Reset CloudWatch alarm.")

	if err := cloudwatchAPI.UpdateKinesisUpscaleAlarm(ctx, stream+"_upscale", stream, topicArn, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", stream+"_upscale").WithField("stream", stream).WithField("count", newShards).Debug("Reset CloudWatch alarm.")

	return nil
}

func downscale(shards float64) int64 {
	switch {
	case shards < 5:
		return int64(math.Ceil(shards / 2))
	case shards < 13:
		return int64(math.Ceil(shards / 1.75))
	case shards < 33:
		return int64(math.Ceil(shards / 1.5))
	default:
		return int64(math.Ceil(shards / 1.25))
	}
}

func upscale(shards float64) int64 {
	switch {
	case shards < 5:
		return int64(math.Floor(shards * 2))
	case shards < 13:
		return int64(math.Floor(shards * 1.75))
	case shards < 33:
		return int64(math.Floor(shards * 1.5))
	default:
		return int64(math.Floor(shards * 1.25))
	}
}
