package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brexhq/substation/internal/aws/cloudwatch"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/json"
	"github.com/brexhq/substation/internal/log"
)

const (
	autoscalePercentage = 50.0
)

var (
	cloudwatchAPI cloudwatch.API
	kinesisAPI    kinesis.API
)

func init() {
	cloudwatchAPI.Setup()
	kinesisAPI.Setup()
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, snsEvent events.SNSEvent) error {
	payload := snsEvent.Records[0].SNS
	topicArn := payload.TopicArn
	message := payload.Message

	alarmName := json.Get([]byte(message), "AlarmName").String()
	triggerMetrics := json.Get([]byte(message), "Trigger.Metrics")

	log.WithField("alarm", alarmName).Info("received autoscale notification")

	var stream string
	for _, v := range triggerMetrics.Array() {
		id := json.Get([]byte(v.String()), "Id").String()
		if id == "m1" || id == "m2" {
			stream = json.Get([]byte(v.String()), "MetricStat.Metric.Dimensions.0.value").String()
			break
		}
	}
	log.WithField("alarm", alarmName).WithField("stream", stream).Info("parsed Kinesis stream")

	shards, err := kinesisAPI.ActiveShards(ctx, stream)
	if err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", shards).
		Info("retrieved active shard count")

	var newShards int64
	if strings.Contains(alarmName, "upscale") {
		newShards = upscale(float64(shards), autoscalePercentage)
	}
	if strings.Contains(alarmName, "downscale") {
		newShards = downscale(float64(shards), autoscalePercentage)
	}

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

			log.WithField("stream", stream).WithField("count", minShard).Info("retrieved minimum shard count")
		}

		if *tag.Key == "MaximumShards" {
			maxShard, err = strconv.ParseInt(*tag.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("handler: %v", err)
			}

			log.WithField("stream", stream).WithField("count", maxShard).Info("retrieved maximum shard count")
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
		log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", shards).Info("active shard count is at minimum threshold, no updates necessary")
		return nil
	}

	if err := kinesisAPI.UpdateShards(ctx, stream, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", alarmName).WithField("stream", stream).WithField("count", newShards).Info("updated shards")

	if err := cloudwatchAPI.UpdateKinesisDownscaleAlarm(ctx, stream+"_downscale", stream, topicArn, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", stream+"_downscale").WithField("stream", stream).WithField("count", newShards).Info("reset alarm")

	if err := cloudwatchAPI.UpdateKinesisUpscaleAlarm(ctx, stream+"_upscale", stream, topicArn, newShards); err != nil {
		return fmt.Errorf("handler: %v", err)
	}
	log.WithField("alarm", stream+"_upscale").WithField("stream", stream).WithField("count", newShards).Info("reset alarm")

	return nil
}

func downscale(shards, pct float64) int64 {
	return int64(math.Ceil(shards - (shards * (pct / 100))))
}

func upscale(shards, pct float64) int64 {
	return int64(math.Ceil(shards + (shards * (pct / 100))))
}
