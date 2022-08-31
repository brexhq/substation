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
)

var sub cmd.Substation

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
	ModificationTime time.Time `json:"modification_time"`
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
	concurrency, err := cmd.GetConcurrency()
	if err != nil {
		return fmt.Errorf("file concurrency: %v", err)
	}

	sub.CreateChannels(concurrency)
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		var transformWg sync.WaitGroup
		for w := 1; w <= concurrency; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		fileHandle, err := os.Open(filename)
		if err != nil {
			sub.SendErr(fmt.Errorf("file filename %s: %v", filename, err))
			return
		}
		defer fileHandle.Close()

		fi, err := fileHandle.Stat()
		if err != nil {
			sub.SendErr(fmt.Errorf("file filename %s: %v", filename, err))
			return
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

		for scanner.Scan() {
			cap.SetData([]byte(scanner.Text()))
			sub.SendTransform(cap)
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return fmt.Errorf("file: %v", err)
	}

	return nil
}
