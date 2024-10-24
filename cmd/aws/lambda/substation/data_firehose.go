package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
)

type firehoseMetadata struct {
	ApproximateArrivalTimestamp time.Time `json:"approximateArrivalTimestamp"`
	RecordId                    string    `json:"recordId"`
}

func firehoseHandler(ctx context.Context, event events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	var resp events.KinesisFirehoseResponse

	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return resp, err
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return resp, err
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return resp, err
	}

	// Records are transformed individually. Firehose cannot produce multiple records
	// from a single record, so the first transformed message is used and the rest
	// are dropped. If no messages are produced, then the record is "dropped."
	for _, record := range event.Records {
		m := firehoseMetadata{
			ApproximateArrivalTimestamp: record.ApproximateArrivalTimestamp.Time,
			RecordId:                    record.RecordID,
		}
		metadata, err := json.Marshal(m)
		if err != nil {
			return resp, err
		}

		msg := message.New().SetData(record.Data).SetMetadata(metadata).SkipMissingValues()
		res, err := sub.Transform(ctx, msg)
		if err != nil {
			return resp, err
		}

		if len(res) == 0 {
			resp.Records = append(resp.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateDropped,
			})
		} else {
			resp.Records = append(resp.Records, events.KinesisFirehoseResponseRecord{
				RecordID: record.RecordID,
				Result:   events.KinesisFirehoseTransformedStateOk,
				Data:     res[0].Data(),
			})
		}
	}

	return resp, nil
}
