package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aws/dynamodb"
	"golang.org/x/sync/errgroup"
)

type dynamodbMetadata struct {
	ApproximateCreationDateTime time.Time `json:"approximateCreationDateTime"`
	EventSourceArn              string    `json:"eventSourceArn"`
	SequenceNumber              string    `json:"sequenceNumber"`
	SizeBytes                   int64     `json:"sizeBytes"`
	StreamViewType              string    `json:"streamViewType"`
}

// nolint: gocognit // ignore cognitive complexity
func dynamodbHandler(ctx context.Context, event events.DynamoDBEvent) error {
	sub := cmd.New()

	// retrieve and load configuration
	cfg, err := getConfig(ctx)
	if err != nil {
		return fmt.Errorf("dynamodb handler: %v", err)
	}

	if err := sub.SetConfig(cfg); err != nil {
		return fmt.Errorf("dynamodb handler: %v", err)
	}

	// maintains app state
	group, ctx := errgroup.WithContext(ctx)

	// load
	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	// transform
	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	group.Go(func() error {
		// the DynamoDB table name is the second element of the slash-delimited Stream ARN.
		// arn:aws:dynamodb:us-west-2:111122223333:table/TestTable/stream/2015-05-11T21:21:33.291
		table := strings.Split(event.Records[0].EventSourceArn, "/")[1]

		for _, record := range event.Records {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// only records that contain image data (changes) are supported.
				if record.Change.StreamViewType == "KEYS_ONLY" {
					continue
				}

				// DynamoDB record changes are converted to an object modeled similarly to
				// schemas used in Debezium (https://debezium.io/). if the View Type on the
				// Stream is OLD_IMAGE, then the "after" field is always null; if the View
				// Type is NEW_IMAGE, then the "before" field is always null. setting the
				// View Type to NEW_AND_OLD_IMAGES is recommended.
				//
				// for more information see these examples from the Debezium documentation:
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
				capsule := config.NewCapsule()

				if err := capsule.Set("source.ts_ms", record.Change.ApproximateCreationDateTime.Time.UnixMilli()); err != nil {
					return fmt.Errorf("dynamodb handler: %v", err)
				}

				if err := capsule.Set("source.table", table); err != nil {
					return fmt.Errorf("dynamodb handler: %v", err)
				}

				if err := capsule.Set("source.connector", "dynamodb"); err != nil {
					return fmt.Errorf("dynamodb handler: %v", err)
				}

				if err := capsule.Set("ts_ms", time.Now().UnixMilli()); err != nil {
					return fmt.Errorf("dynamodb handler: %v", err)
				}

				// maps the type of data modification to a Debezium operation string.
				// Debezium operations that are relevant to DynamoDB are:
				// - c: create (INSERT)
				// - u: update (MODIFY)
				// - d: delete (REMOVE)
				switch record.EventName {
				case "INSERT":
					if err := capsule.Set("op", "c"); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				case "MODIFY":
					if err := capsule.Set("op", "u"); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				case "REMOVE":
					if err := capsule.Set("op", "d"); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				}

				// if either image is missing, then the value is set to null.
				if record.Change.OldImage == nil {
					if err := capsule.Set("before", nil); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				} else {
					var before map[string]interface{}
					if err = dynamodbattribute.UnmarshalMap(
						dynamodb.ConvertEventsAttributeValueMap(record.Change.OldImage),
						&before,
					); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}

					if err := capsule.Set("before", before); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				}

				if record.Change.NewImage == nil {
					if err := capsule.Set("after", nil); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				} else {
					var after map[string]interface{}
					if err = dynamodbattribute.UnmarshalMap(
						dynamodb.ConvertEventsAttributeValueMap(record.Change.NewImage),
						&after,
					); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}

					if err := capsule.Set("after", after); err != nil {
						return fmt.Errorf("dynamodb handler: %v", err)
					}
				}

				if _, err := capsule.SetMetadata(dynamodbMetadata{
					record.Change.ApproximateCreationDateTime.Time,
					record.EventSourceArn,
					record.Change.SequenceNumber,
					record.Change.SizeBytes,
					record.Change.StreamViewType,
				}); err != nil {
					return fmt.Errorf("dynamodb handler: %v", err)
				}

				sub.Send(capsule)
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	// block until ITL is complete
	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("dynamodb handler: %v", err)
	}

	return nil
}
