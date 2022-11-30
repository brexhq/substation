package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/file"
	"golang.org/x/sync/errgroup"
)

type metadata struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
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
	input := flag.String("input", "", "file to parse")
	config := flag.String("config", "", "Substation configuration file")
	timeout := flag.Duration("timeout", 10*time.Second, "timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := run(ctx, *input, *config); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

func run(ctx context.Context, input, cfg string) error {
	sub := cmd.New()

	// load configuration file
	c, err := getConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("run: %v", err)
	}

	if err := sub.SetConfig(c); err != nil {
		return fmt.Errorf("run: %v", err)
	}

	group, ctx := errgroup.WithContext(ctx)

	var sinkWg sync.WaitGroup
	sinkWg.Add(1)
	group.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	var transformWg sync.WaitGroup
	for w := 0; w < sub.Concurrency(); w++ {
		transformWg.Add(1)
		group.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	// ingest
	group.Go(func() error {
		fi, err := file.Get(ctx, input)
		if err != nil {
			return err
		}
		defer os.Remove(fi)

		f, err := os.Open(fi)
		if err != nil {
			return fmt.Errorf("run: %v", err)
		}
		defer f.Close()

		fs, err := f.Stat()
		if err != nil {
			return err
		}

		capsule := config.NewCapsule()
		if _, err = capsule.SetMetadata(metadata{
			input,
			fs.Size(),
		}); err != nil {
			return fmt.Errorf("run: %v", err)
		}

		scanner := bufio.NewScanner()
		defer scanner.Close()

		scanner.ReadFile(f)
		for scanner.Scan() {
			switch scanner.Method() {
			case "bytes":
				capsule.SetData(scanner.Bytes())
			case "text":
				capsule.SetData([]byte(scanner.Text()))
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				sub.Send(capsule)
			}
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return fmt.Errorf("run: %v", err)
	}

	return nil
}
