package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/awslabs/kinesis-aggregation/go/v2/deaggregator"
	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/channel"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/file"
	"github.com/brexhq/substation/v2/internal/log"
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

	flag.StringVar(&opts.Config, "config", "./config.json", "The Substation configuration file used to transform records")
	flag.StringVar(&opts.StreamName, "stream-name", "", "The AWS Kinesis Data Stream to fetch records from")
	flag.StringVar(&opts.StreamOffset, "stream-offset", "earliest", "Determines the offset of the stream (earliest, latest)")
	flag.Parse()

	if err := run(context.Background(), opts); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

//nolint:gocognit // Ignore cognitive complexity.
func run(ctx context.Context, opts options) error {
	cfg := substation.Config{}
	c, err := getConfig(ctx, opts.Config)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(c).Decode(&cfg); err != nil {
		return err
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return err
	}

	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	// Consumer group that transforms records using Substation
	// until the channel is closed by the producer group.
	group.Go(func() error {
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

		log.Debug("Closed Substation pipeline.")

		// ctrl messages flush the pipeline. This must be done
		// after all messages have been processed.
		ctrl := message.New().AsControl()
		if _, err := sub.Transform(tfCtx, ctrl); err != nil {
			return err
		}

		log.Debug("Flushed Substation pipeline.")

		return nil
	})

	// Producer group that fetches records from each shard in the
	// Kinesis stream until the context is cancelled by an interrupt
	// signal.
	group.Go(func() error {
		defer ch.Close() // Producer goroutines must close the channel when they are done.

		// The AWS client is configured using environment variables
		// or the default credentials file.
		awsCfg, err := iconfig.NewAWS(ctx, iconfig.AWS{})
		if err != nil {
			return err
		}

		client := kinesis.NewFromConfig(awsCfg)

		resp, err := client.ListShards(ctx, &kinesis.ListShardsInput{
			StreamName: &opts.StreamName,
		})
		if err != nil {
			return err
		}

		log.WithField("stream", opts.StreamName).WithField("count", len(resp.Shards)).Debug("Retrieved active shards from Kinesis stream.")

		var iType string
		switch opts.StreamOffset {
		case "earliest":
			iType = "TRIM_HORIZON"
		case "latest":
			iType = "LATEST"
		default:
			return fmt.Errorf("invalid offset: %s", opts.StreamOffset)
		}

		// Each shard is read concurrently using a worker
		// pool managed by an errgroup that can be cancelled
		// by an interrupt signal.
		notifyCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT)
		defer cancel()

		recvGroup, recvCtx := errgroup.WithContext(notifyCtx)
		defer log.Debug("Closed connections to the Kinesis stream.")

		// This iterates over a snapshot of active shards in the
		// stream and will not be updated if shards are split or
		// merged. New shards can be identified in the response
		// from GetRecords, but this isn't implemented.
		//
		// Each shard is paginated until the end of the shard is
		// reached or the context is cancelled.
		for _, shard := range resp.Shards {
			iterator, err := client.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
				StreamName:        &opts.StreamName,
				ShardId:           shard.ShardId,
				ShardIteratorType: types.ShardIteratorType(iType),
			})
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

					// GetRecords has a limit of 5 transactions per second
					// per shard, so this loop is designed to not overload
					// the API in case other consumers are reading from the
					// same shard.
					resp, err := client.GetRecords(recvCtx, &kinesis.GetRecordsInput{
						ShardIterator: &shardIterator,
					})
					if err != nil {
						return err
					}

					if resp.NextShardIterator == nil {
						log.WithField("stream", opts.StreamName).WithField("shard", shard.ShardId).Debug("Reached end of Kinesis shard.")

						break
					}
					shardIterator = *resp.NextShardIterator

					if len(resp.Records) == 0 {
						time.Sleep(500 * time.Millisecond)

						continue
					}

					deagg, err := deaggregator.DeaggregateRecords(resp.Records)
					if err != nil {
						return err
					}

					log.WithField("stream", opts.StreamName).WithField("shard", shard.ShardId).WithField("count", len(deagg)).Debug("Retrieved records from Kinesis shard.")

					for _, record := range deagg {
						msg := message.New().SetData(record.Data).SkipMissingValues()
						ch.Send(msg)
					}

					time.Sleep(500 * time.Millisecond)
				}

				return nil
			})
		}

		// Cancellation errors are expected when the errgroup
		// is interrupted. All other errors are returned to
		// the caller.
		if err := recvGroup.Wait(); err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "RequestCanceled" {
					return nil
				}
			}

			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		}

		return nil
	})

	// Wait for the producer and consumer groups to finish,
	// or an error from either group.
	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}
