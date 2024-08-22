package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/message"
	"golang.org/x/sync/errgroup"
)

type dynamodbMetadata struct {
	ApproximateCreationDateTime time.Time `json:"approximateCreationDateTime"`
	EventSourceArn              string    `json:"eventSourceArn"`
	SequenceNumber              string    `json:"sequenceNumber"`
	SizeBytes                   int64     `json:"sizeBytes"`
	StreamViewType              string    `json:"streamViewType"`
}

//nolint:gocognit, gocyclo, cyclop // Ignore cognitive and cyclomatic complexity.
func dynamodbHandler(ctx context.Context, event events.DynamoDBEvent) error {
	// Retrieve and load configuration.
	conf, err := getConfig(ctx)
	if err != nil {
		return err
	}

	cfg := customConfig{}
	if err := json.NewDecoder(conf).Decode(&cfg); err != nil {
		return err
	}

	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return err
	}

	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Data transformation. Transforms are executed concurrently using a worker pool
	// managed by an errgroup. Each message is processed in a separate goroutine.
	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(cfg.Concurrency)

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			msg := message
			tfGroup.Go(func() error {
				// Transformed messages are never returned to the caller because
				// invocation is asynchronous.
				if _, err := sub.Transform(tfCtx, msg); err != nil {
					return err
				}

				return nil
			})
		}

		if err := tfGroup.Wait(); err != nil {
			return err
		}

		// CTRL messages flush the pipeline. This must be done
		// after all messages have been processed.
		ctrl := message.New().AsControl()
		if _, err := sub.Transform(tfCtx, ctrl); err != nil {
			return err
		}

		return nil
	})

	// Data ingest.
	group.Go(func() error {
		defer ch.Close()

		// The DynamoDB table name is the second element of the slash-delimited Stream ARN.
		// arn:aws:dynamodb:us-west-2:111122223333:table/TestTable/stream/2015-05-11T21:21:33.291
		table := strings.Split(event.Records[0].EventSourceArn, "/")[1]

		for _, record := range event.Records {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Only records that contain image data (changes) are supported.
			if record.Change.StreamViewType == "KEYS_ONLY" {
				continue
			}

			m := dynamodbMetadata{
				record.Change.ApproximateCreationDateTime.Time,
				record.EventSourceArn,
				record.Change.SequenceNumber,
				record.Change.SizeBytes,
				record.Change.StreamViewType,
			}
			metadata, err := json.Marshal(m)
			if err != nil {
				return err
			}

			// DynamoDB record changes are converted to an object modeled similarly to
			// schemas used in Debezium (https://debezium.io/):
			//
			// - If the View Type on the Stream is OLD_IMAGE, then the "after" field is always null.
			// - If the View Type is NEW_IMAGE, then the "before" field is always null.
			//
			// Setting the View Type to NEW_AND_OLD_IMAGES is recommended for full visibility.
			//
			// For more information, see these examples from the Debezium documentation:
			// - https://debezium.io/documentation/reference/1.2/connectors/mysql.html#mysql-change-event-value
			// - https://debezium.io/documentation/reference/1.2/connectors/postgresql.html#postgresql-change-event-value
			// - https://debezium.io/documentation/reference/1.2/connectors/sqlserver.html#sqlserver-change-event-value
			//
			// records are converted to this format:
			// {
			//   "source": {
			//     "ts_ms": 0,
			//     "table": "table",
			//     "connector": "dynamodb"
			//   },
			//   "ts_ms": 0,
			//   "op": "c",
			//   "before": { ... },
			//   "after": { ... }
			// }
			msg := message.New().SetMetadata(metadata)
			if err := msg.SetValue("source.ts_ms", record.Change.ApproximateCreationDateTime.Time.UnixMilli()); err != nil {
				return err
			}

			if err := msg.SetValue("source.table", table); err != nil {
				return err
			}

			if err := msg.SetValue("source.connector", "dynamodb"); err != nil {
				return err
			}

			if err := msg.SetValue("ts_ms", time.Now().UnixMilli()); err != nil {
				return err
			}

			// Maps the type of data modification to a Debezium operation string.
			// Debezium operations that are relevant to DynamoDB are:
			// - c: create (INSERT)
			// - u: update (MODIFY)
			// - d: delete (REMOVE)
			switch record.EventName {
			case "INSERT":
				if err := msg.SetValue("op", "c"); err != nil {
					return err
				}
			case "MODIFY":
				if err := msg.SetValue("op", "u"); err != nil {
					return err
				}
			case "REMOVE":
				if err := msg.SetValue("op", "d"); err != nil {
					return err
				}
			}

			// If either image is missing, then the value is set to null.
			if record.Change.OldImage == nil {
				if err := msg.SetValue("before", nil); err != nil {
					return err
				}
			} else {
				var before map[string]interface{}
				if err = dynamodbattribute.UnmarshalMap(
					dynamodb.ConvertEventsAttributeValueMap(record.Change.OldImage),
					&before,
				); err != nil {
					return err
				}

				if err := msg.SetValue("before", before); err != nil {
					return err
				}
			}

			if record.Change.NewImage == nil {
				if err := msg.SetValue("after", nil); err != nil {
					return err
				}
			} else {
				var after map[string]interface{}
				if err = dynamodbattribute.UnmarshalMap(
					dynamodb.ConvertEventsAttributeValueMap(record.Change.NewImage),
					&after,
				); err != nil {
					return err
				}

				if err := msg.SetValue("after", after); err != nil {
					return err
				}
			}

			ch.Send(msg)
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}
