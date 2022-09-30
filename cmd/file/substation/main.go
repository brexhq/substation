package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"golang.org/x/sync/errgroup"
)

var sub cmd.Substation
var scanMethod string

func loadConfig(f string) error {
	bytes, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	json.Unmarshal(bytes, &sub.Config)

	return nil
}

type metadata struct {
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	ModificationTime time.Time `json:"modificationTime"`
}

func main() {
	input := flag.String("input", "", "file to parse from local disk")
	config := flag.String("config", "", "Substation configuration file")
	timeout := flag.Duration("timeout", 10*time.Second, "timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	loadConfig(*config)

	if err := file(ctx, *input); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

func file(ctx context.Context, filename string) error {
	defer fmt.Println(runtime.NumGoroutine())

	// retrieves concurrency value from SUBSTATION_CONCURRENCY environment variable
	concurrency, err := cmd.GetConcurrency()
	if err != nil {
		return fmt.Errorf("file concurrency: %v", err)
	}

	// retrieves scan method from SUBSTATION_SCAN_METHOD environment variable
	scanMethod = cmd.GetScanMethod()

	sub.CreateChannels(concurrency)

	g, ctx := errgroup.WithContext(ctx)

	// this seems bad, but it enforces panics in any goroutine producer that is blocked due to a dead consumer
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			if ctx.Err() != nil {
				sub.CloseChannels()
			}
		}
	}()

	var sinkWg sync.WaitGroup
	var transformWg sync.WaitGroup

	// sink
	sinkWg.Add(1)
	g.Go(func() error {
		return sub.Sink(ctx, &sinkWg)
	})

	// transforms
	for w := 0; w < concurrency; w++ {
		transformWg.Add(1)
		g.Go(func() error {
			return sub.Transform(ctx, &transformWg)
		})
	}

	g.Go(func() error {
		fileHandle, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer fileHandle.Close()

		fi, err := fileHandle.Stat()
		if err != nil {
			return err
		}

		cap := config.NewCapsule()
		cap.SetMetadata(metadata{
			fi.Name(),
			fi.Size(),
			fi.ModTime(),
		})

		// a scanner token can be up to 100MB
		scanner := bufio.NewScanner(fileHandle)
		scanner.Buffer([]byte{}, 100*1024*1024)

		var count int
		for scanner.Scan() {
			switch scanMethod {
			case "bytes":
				cap.SetData(scanner.Bytes())
			case "text":
				cap.SetData([]byte(scanner.Text()))
			}

			select {
			case <-ctx.Done():
				return nil
			default:
				sub.SendTransform(cap)
			}

			count++
		}

		sub.TransformWait(&transformWg)
		sub.SinkWait(&sinkWg)

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
