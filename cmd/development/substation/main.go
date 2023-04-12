package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/file"
	"github.com/brexhq/substation/internal/json"
)

type metadata struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type options struct {
	Input  string
	Config string

	ForceSink string
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

	timeout := flag.Duration("timeout", 10*time.Second, "timeout")
	flag.StringVar(&opts.Input, "input", "", "file to parse")
	flag.StringVar(&opts.Config, "config", "", "Substation configuration file")
	flag.StringVar(&opts.ForceSink, "force-sink", "", "force sink output to value (supported: stdout)")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := run(ctx, opts); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

func run(ctx context.Context, opts options) error {
	sub := cmd.New()

	// load configuration file
	c, err := getConfig(ctx, opts.Config)
	if err != nil {
		return fmt.Errorf("run: %v", err)
	}

	if err := sub.SetConfig(c); err != nil {
		return fmt.Errorf("run: %v", err)
	}

	if opts.ForceSink != "" {
		c, err = sub.Config()
		if err != nil {
			return fmt.Errorf("run: %v", err)
		}

		newConfig, err := mutateSink(c, opts.ForceSink)
		if err != nil {
			return fmt.Errorf("run: %v", err)
		}

		if err := sub.SetConfig(newConfig); err != nil {
			return fmt.Errorf("run: %v", err)
		}
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
		fi, err := file.Get(ctx, opts.Input)
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
			opts.Input,
			fs.Size(),
		}); err != nil {
			return fmt.Errorf("run: %v", err)
		}

		scanner := bufio.NewScanner()
		defer scanner.Close()

		if err := scanner.ReadFile(f); err != nil {
			return fmt.Errorf("run: %v", err)
		}

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

func mutateSink(cfg io.Reader, forceSink string) (*bytes.Reader, error) {
	oldConfig, err := io.ReadAll(cfg)
	if err != nil {
		return nil, fmt.Errorf("run: %v", err)
	}

	// removes the configured sink
	oldConfig, err = json.Delete(oldConfig, "sink")
	if err != nil {
		return nil, fmt.Errorf("run: %v", err)
	}

	var newSink string

	switch {
	case forceSink == "stdout" || forceSink == "file":
		newSink = fmt.Sprintf(`{"type": "%s"}`, forceSink)
	case strings.HasPrefix(forceSink, "file://"):
		newSink = fmt.Sprintf(
			`{"type": "file", "settings": {"file_path": {"suffix": "%s"}}}`,
			strings.TrimPrefix(forceSink, "file://"),
		)
	case strings.HasPrefix(forceSink, "http://") || strings.HasPrefix(forceSink, "https://"):
		newSink = fmt.Sprintf(
			`{"type": "http", "settings": {"url": "%s"}}`,
			forceSink,
		)
	case strings.HasPrefix(forceSink, "s3://"):
		return nil, fmt.Errorf("-force-sink s3://* not yet implemented")
	default:
		return nil, fmt.Errorf("%q not supported for -force-sink", forceSink)
	}

	newConfig, err := json.Set(oldConfig, "sink", newSink)
	if err != nil {
		return nil, fmt.Errorf("run: %v", err)
	}

	return bytes.NewReader(newConfig), nil
}
