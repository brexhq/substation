package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/aws/smithy-go"
	"github.com/awslabs/kinesis-aggregation/go/v2/deaggregator"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/channel"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/log"
)

func init() {
	rootCmd.AddCommand(tapCmd)
	tapCmd.PersistentFlags().String("aws-kinesis-data-stream", "", "arn of the aws kinesis data stream to tap")
	tapCmd.PersistentFlags().String("offset", "latest", "the offset to read from (earliest, latest)")
	tapCmd.PersistentFlags().StringToString("ext-str", nil, "set external variables")
	tapCmd.Flags().SortFlags = false
	tapCmd.PersistentFlags().SortFlags = false
}

var tapCmd = &cobra.Command{
	Use:   "tap [path]",
	Short: "tap data streams",
	Long: `'substation tap' reads from a data stream.
It supports these data stream sources:
  AWS Kinesis Data Streams (--aws-kinesis-data-stream)

The data stream can be read from either the beginning 
(earliest) or the end (latest) using the --offset flag.
Reading the stream can be interrupted by sending an 
interrupt signal (ex. Ctrl+C).

If the config is not already compiled, then it is compiled 
before reading the stream ('.jsonnet', '.libsonnet' files are 
compiled to JSON). If no config is provided, then the stream
data is sent to stdout.

Debug logs can be enabled to report the status of reading
from the data stream. Use this environment variable to
enable debug logs: SUBSTATION_DEBUG=true

WARNING: This command is intended to provide temporary access 
to streaming data and should not be used for production workloads.

WARNING: This command is "experimental" and does not strictly 
adhere to semantic versioning. Refer to the versioning policy
for more information.
`,
	Example: `  substation tap --aws-kinesis-data-stream arn:aws:kinesis:us-east-1:123456789012:stream/my-stream
  substation tap --aws-kinesis-data-stream arn:aws:kinesis:us-east-1:123456789012:stream/my-stream --offset earliest
  substation tap /path/to/config.json --aws-kinesis-data-stream arn:aws:kinesis:us-east-1:123456789012:stream/my-stream
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no path is provided, then a default config is used.
		path := ""
		if len(args) > 0 {
			path = args[0]
		}

		// Catches an edge case where the user is looking for help.
		if path == "help" {
			fmt.Printf("warning: use -h instead.\n")
			return nil
		}

		ext, err := cmd.PersistentFlags().GetStringToString("ext-str")
		if err != nil {
			return err
		}

		offset, err := cmd.Flags().GetString("offset")
		if err != nil {
			return err
		}

		kinesis, err := cmd.Flags().GetString("aws-kinesis-data-stream")
		if err != nil {
			return err
		}

		if kinesis != "" {
			return tapKinesis(path, ext, offset, kinesis)
		}

		return fmt.Errorf("no valid data stream source provided")
	},
}

type kinesisStreamMetadata struct {
	ApproximateArrivalTimestamp time.Time `json:"approximateArrivalTimestamp"`
	Stream                      string    `json:"stream"`
	PartitionKey                string    `json:"partitionKey"`
	SequenceNumber              string    `json:"sequenceNumber"`
}

//nolint:gocognit, cyclop, gocyclo // Ignore cognitive and cyclomatic complexity.
func tapKinesis(arg string, extVars map[string]string, offset, stream string) error {
	cfg := customConfig{}

	switch filepath.Ext(arg) {
	case ".jsonnet", ".libsonnet":
		mem, err := compileFile(arg, extVars)
		if err != nil {
			// This is an error in the Jsonnet syntax.
			// The line number and column range are included.
			//
			// Example: `vet.jsonnet:19:36-38 Unknown variable: st`
			fmt.Printf("%v\n", err)

			return nil
		}

		cfg, err = memConfig(mem)
		if err != nil {
			return err
		}
	case ".json":
		fi, err := fiConfig(arg)
		if err != nil {
			return err
		}

		cfg = fi
	default:
		mem, err := compileStr(confStdout, extVars)
		if err != nil {
			return err
		}

		cfg, err = memConfig(mem)
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	sub, err := substation.New(ctx, cfg.Config)
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
			StreamARN: &stream,
		})
		if err != nil {
			return err
		}

		log.WithField("stream", stream).WithField("count", len(resp.Shards)).Debug("Retrieved active shards from Kinesis stream.")

		var iType string
		switch offset {
		case "earliest":
			iType = "TRIM_HORIZON"
		case "latest":
			iType = "LATEST"
		default:
			return fmt.Errorf("invalid offset: %s", stream)
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
				StreamARN:         &stream,
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
						log.WithField("stream", stream).WithField("shard", shard.ShardId).Debug("Reached end of Kinesis shard.")

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

					log.WithField("stream", stream).WithField("shard", shard.ShardId).WithField("count", len(deagg)).Debug("Retrieved records from Kinesis shard.")

					for _, record := range deagg {
						// Create Message metadata.
						m := kinesisStreamMetadata{
							*record.ApproximateArrivalTimestamp,
							stream,
							*record.PartitionKey,
							*record.SequenceNumber,
						}
						metadata, err := json.Marshal(m)
						if err != nil {
							return err
						}

						msg := message.New().SetData(record.Data).SetMetadata(metadata)
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
			var ae smithy.APIError
			if errors.As(err, &ae) {
				if ae.ErrorCode() == "RequestCanceled" {
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
