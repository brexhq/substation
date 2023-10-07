package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
)

type options struct {
	File   string
	Config string
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

	timeout := flag.Duration("timeout", 10*time.Second, "Timeout in seconds")
	flag.StringVar(&opts.File, "file", "", "File to parse")
	flag.StringVar(&opts.Config, "config", "", "Substation configuration file")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := run(ctx, opts); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

func run(ctx context.Context, opts options) error {
	c, err := getConfig(ctx, opts.Config)
	if err != nil {
		return err
	}

	cfg := substation.Config{}
	if err := json.NewDecoder(c).Decode(&cfg); err != nil {
		return err
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		return err
	}

	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(runtime.NumCPU())

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			message := message
			tfGroup.Go(func() error {
				if _, err := transform.Apply(tfCtx, sub.Transforms(), message); err != nil {
					return err
				}

				return nil
			})
		}

		if err := tfGroup.Wait(); err != nil {
			return err
		}

		// CTRL Messages flush the transform functions. This must be done
		// after all messages have been processed.
		ctrl := message.New(message.AsControl())
		if _, err := transform.Apply(ctx, sub.Transforms(), ctrl); err != nil {
			return err
		}

		return nil
	})

	group.Go(func() error {
		defer ch.Close()

		fi, err := file.Get(ctx, opts.File)
		if err != nil {
			return err
		}
		defer os.Remove(fi)

		f, err := os.Open(fi)
		if err != nil {
			return err
		}
		defer f.Close()

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
			msg := message.New().SetData(b)

			ch.Send(msg)
		}

		if err := scanner.Err(); err != nil {
			return err
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
