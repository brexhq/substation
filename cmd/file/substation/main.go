package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"golang.org/x/sync/errgroup"
)

var sub *cmd.Substation

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

	sub = cmd.New()
	loadConfig(*config)

	if err := file(ctx, *input); err != nil {
		panic(fmt.Errorf("main: %v", err))
	}
}

func file(ctx context.Context, filename string) error {
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

		// retrieves scan method from SUBSTATION_SCAN_METHOD environment variable
		scanMethod := cmd.GetScanMethod()

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
				return ctx.Err()
			default:
				sub.Send(cap)
			}

			count++
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		return err
	}

	return nil
}
