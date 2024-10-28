package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/bufio"
	"github.com/brexhq/substation/v2/internal/channel"
	iconfig "github.com/brexhq/substation/v2/internal/config"
	"github.com/brexhq/substation/v2/internal/media"
)

func init() {
	rootCmd.AddCommand(readCmd)
	readCmd.PersistentFlags().String("file", "", "path to the file to read")
	readCmd.PersistentFlags().String("http", "", "http(s) endpoint to read")
	readCmd.PersistentFlags().String("aws", "", "aws s3 object to read")
	readCmd.PersistentFlags().StringToString("ext-str", nil, "set external variables")
	readCmd.Flags().SortFlags = false
	readCmd.PersistentFlags().SortFlags = false
}

var readCmd = &cobra.Command{
	Use:   "read [path]",
	Short: "read files",
	Long: `'substation read' reads data from a file.
It supports these file sources:
  Local File (--file)
  HTTP(S) Endpoint (--http)
  AWS S3 Object (--aws)

If the config is not already compiled, then it is compiled 
before reading the stream ('.jsonnet', '.libsonnet' files are 
compiled to JSON). If no config is provided, then the stream
data is sent to stdout.

WARNING: This command is "experimental" and does not strictly 
adhere to semantic versioning. Refer to the versioning policy
for more information.
`,
	Example: `  substation read --file /path/to/file.json
  substation read --http https://example.com
  substation read --aws s3://bucket/path/to/file.json
  substation read /path/to/config.json --file /path/to/file.json
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

		var cfg customConfig

		switch filepath.Ext(path) {
		case ".jsonnet", ".libsonnet":
			mem, err := compileFile(path, ext)
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
			fi, err := fiConfig(path)
			if err != nil {
				return err
			}

			cfg = fi
		default:
			mem, err := compileStr(confStdout, ext)
			if err != nil {
				return err
			}

			cfg, err = memConfig(mem)
			if err != nil {
				return err
			}
		}

		switch {
		case cmd.Flags().Lookup("file").Changed:
			fi, err := cmd.PersistentFlags().GetString("file")
			if err != nil {
				return err
			}

			f, err := readFile(fi)
			if err != nil {
				return err
			}

			return read(cfg, f)
		case cmd.Flags().Lookup("http").Changed:
			fi, err := cmd.PersistentFlags().GetString("http")
			if err != nil {
				return err
			}

			f, err := readHTTP(fi)
			defer func() { // Always clean up the temp file.
				_ = f.Close()
				_ = os.Remove(f.Name())
			}()

			if err != nil {
				return err
			}

			if err := read(cfg, f); err != nil {
				return err
			}

			return nil
		case cmd.Flags().Lookup("aws").Changed:
			fi, err := cmd.PersistentFlags().GetString("aws")
			if err != nil {
				return err
			}

			f, err := readS3(fi)
			defer func() { // Always clean up the temp file.
				_ = f.Close()
				_ = os.Remove(f.Name())
			}()

			if err != nil {
				return err
			}

			if err := read(cfg, f); err != nil {
				return err
			}

			return nil
		}

		return fmt.Errorf("no valid file source provided")
	},
}

func read(cfg customConfig, f *os.File) error {
	if f == nil {
		return fmt.Errorf("invalid file")
	}

	ctx := context.Background()
	sub, err := substation.New(ctx, cfg.Config)
	if err != nil {
		return err
	}

	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(1) // Set to 1 to process messages sequentially.

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

	group.Go(func() error {
		defer ch.Close()

		mediaType, err := media.File(f)
		if err != nil {
			return err
		}

		if _, err := f.Seek(0, 0); err != nil {
			return err
		}

		// Unsupported media types are sent as binary data.
		if !slices.Contains(bufio.MediaTypes, mediaType) {
			r, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			msg := message.New().SetData(r).SkipMissingValues()
			ch.Send(msg)

			return nil
		}

		scanner := bufio.NewScanner()
		defer scanner.Close()

		if err := scanner.ReadFile(f); err != nil {
			return err
		}

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			b := []byte(scanner.Text())
			msg := message.New().SetData(b).SkipMissingValues()

			ch.Send(msg)
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	})

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func readFile(fi string) (*os.File, error) {
	if _, err := os.Stat(fi); err != nil {
		return nil, fmt.Errorf("invalid file: %s", fi)
	}

	f, err := os.Open(fi)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func readHTTP(fi string) (*os.File, error) {
	f, err := os.CreateTemp("", "substation")
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(fi, "http://") && !strings.HasPrefix(fi, "https://") {
		return f, fmt.Errorf("invalid http endpoint: %s", fi)
	}

	resp, err := retryablehttp.Get(fi)
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()

	size, err := io.Copy(f, resp.Body)
	if err != nil {
		return f, err
	}

	if size == 0 {
		return nil, fmt.Errorf("empty file: %s", fi)
	}

	return f, nil
}

func readS3(fi string) (*os.File, error) {
	f, err := os.CreateTemp("", "substation")
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(fi, "s3://") {
		return f, fmt.Errorf("invalid s3 object: %s", fi)
	}

	ctx := context.Background()
	awsCfg, err := iconfig.NewAWS(ctx, iconfig.AWS{})
	if err != nil {
		return f, err
	}

	c := s3.NewFromConfig(awsCfg)
	s3downloader := manager.NewDownloader(c)

	// "s3://bucket/key" becomes ["bucket" "key"]
	paths := strings.SplitN(strings.TrimPrefix(fi, "s3://"), "/", 2)

	// Download the file from S3.
	ctx = context.WithoutCancel(ctx)
	size, err := s3downloader.Download(ctx, f, &s3.GetObjectInput{
		Bucket: &paths[0],
		Key:    &paths[1],
	})
	if err != nil {
		return f, err
	}

	if size == 0 {
		return f, fmt.Errorf("empty file: %s", fi)
	}

	return f, nil
}
