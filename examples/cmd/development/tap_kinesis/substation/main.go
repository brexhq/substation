// Used as a development tool to test Substation configurations against live data
// by "tapping" an AWS Kinesis stream.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/awslabs/kinesis-aggregation/go/deaggregator"
	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/aws"
	"github.com/brexhq/substation/internal/aws/kinesis"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/internal/log"
	"github.com/brexhq/substation/message"
)

type options struct {
	Config string

	// StreamName is the name of the Kinesis stream to read from.
	StreamName string
	// StreamOffset is the read offset of the stream (earliest, latest).
	StreamOffset string
}

// getConfig contextually retrieves a Substation configuration.
func getConfig(ctx context.Context, cfg string) (io.Reader, error) {
	path, err := file.Get(ctx, cfg)
	defer os.Remove(path)

	if err != nil {
		return nil, err
	}

	conf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer conf.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, conf); err != nil {
		return nil, err
	}

	return buf, nil
}

func main() {
	var opts options

	flag.StringVar(&opts.Config, "config", "", "The Substation configuration file used to transform records")
	flag.StringVar(&opts.StreamName, "stream-name", "", "The AWS Kinesis Data Stream to read records from")
	flag.StringVar(&opts.StreamOffset, "stream-offset", "earliest", "Determines the read offset of the stream (earliest, latest)")
	flag.Parse()

	if err := run(context.Background(), opts); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

func run(ctx context.Context, opts options) error {
	// If no config file is provided, then the app prints Kinesis
	// record data to stdout every 100 records or 5 seconds, whichever
	// happens first.
	cfg := substation.Config{}
	if opts.Config != "" {
		c, err := getConfig(ctx, opts.Config)
		if err != nil {
			return err
		}

		if err := json.NewDecoder(c).Decode(&cfg); err != nil {
			return err
		}
	} else {
		c := []byte(`{"transforms":[
			{
				"type": "send_stdout",
				"settings": {
					"batch": {
						"count": 100,
						"duration": "5s"
					}
				}
			}	
		]}`)
		if err := json.Unmarshal(c, &cfg); err != nil {
			panic(err)
		}
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return err
	}

	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// This group is responsible for transforming records using the Substation
	// configuration. The group finishes when the channel is closed and all
	// messages have been processed, including flushing the pipeline with a ctrl
	// message.
	group.Go(func() error {
		defer log.Info("Closing Substation transforms.")

		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(runtime.NumCPU())

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			msg := message
			tfGroup.Go(func() error {
				if _, err := sub.Transform(tfCtx, msg); err != nil {
					return err
				}

				return nil
			})
		}

		if err := tfGroup.Wait(); err != nil {
			return err
		}

		// ctrl messages flush the pipeline. This must be done
		// after all messages have been processed.
		ctrl := message.New().AsControl()
		if _, err := sub.Transform(tfCtx, ctrl); err != nil {
			return err
		}

		return nil
	})

	// This group is responsible for retrieving records from the Kinesis stream.
	//
	// When the user sends a SIGINT signal (CTRL+C in the terminal) to the app,
	// the workers in this group are interrupted by cancelling the context. This
	// allows the app to gracefully shutdown by continuing to process in-progress
	// messages until the channel is empty (see the transform group above).
	group.Go(func() error {
		defer ch.Close() // Producing goroutines must close the channel when they are done.

		// The client uses settings from environment variables.
		//
		// This can be made configurable in the future.
		client := kinesis.API{}
		client.Setup(aws.Config{})

		ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT)
		defer cancel()

		recvGroup, recvCtx := errgroup.WithContext(ctx)
		defer log.Info("Closing connections to the Kinesis stream.")

		res, err := client.ListShards(ctx, opts.StreamName)
		if err != nil {
			return err
		}

		log.WithField("stream", opts.StreamName).WithField("count", len(res.Shards)).Info("Retrieved active shards from Kinesis stream.")

		// This iterates over a snapshot of active shards in the stream and will not
		// be updated if shard changes (opened, closed) occur. New shards can be identified
		// in the response from GetRecords.
		//
		// Reminder that this is a development tool and is not meant for production use.

		var iType string
		switch opts.StreamOffset {
		case "earliest":
			iType = "TRIM_HORIZON"
		case "latest":
			iType = "LATEST"
		default:
			return fmt.Errorf("invalid offset: %s", opts.StreamOffset)
		}

		for _, shard := range res.Shards {
			iterator, err := client.GetShardIterator(ctx, opts.StreamName, *shard.ShardId, iType)
			if err != nil {
				return err
			}

			recvGroup.Go(func() error {
				shardIterator := *iterator.ShardIterator

				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					res, err := client.GetRecords(recvCtx, shardIterator)
					if err != nil {
						return err
					}

					// This paginates through the shard until it's closed.
					if res.NextShardIterator == nil {
						log.WithField("stream", opts.StreamName).WithField("shard", shard.ShardId).Debug("Reached end of Kinesis shard.")

						break
					} else {
						shardIterator = *res.NextShardIterator
					}

					log.WithField("stream", opts.StreamName).WithField("shard", shard.ShardId).WithField("count", len(res.Records)).Debug("Retrieved records from Kinesis shard.")

					if len(res.Records) == 0 {
						// GetRecords has a limit of 5 transactions per second
						// per shard. This sleep is configured to not overload
						// the API, in case other consumers are reading from the
						// same shard.
						//
						// This can be made configurable in the future.
						time.Sleep(500 * time.Millisecond)

						continue
					}

					deaggregated, err := deaggregator.DeaggregateRecords(res.Records)
					if err != nil {
						return err
					}

					log.WithField("stream", opts.StreamName).WithField("shard", shard.ShardId).WithField("count", len(deaggregated)).Debug("Parsed deaggregated records from Kinesis shard.")

					for _, record := range deaggregated {
						msg := message.New().SetData(record.Data)
						ch.Send(msg)
					}
				}

				return nil
			})
		}

		// AWS errors are expected when the context is cancelled
		// by the user. All other errors are unexpected and returned
		// to the caller.
		if err := recvGroup.Wait(); err != nil {
			if _, ok := err.(awserr.Error); ok {
				return nil
			}

			return err
		}

		return nil
	})

	// Wait for all goroutines to complete.
	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}
