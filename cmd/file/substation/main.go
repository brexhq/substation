package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/brexhq/substation/cmd"
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

func main() {
	input := flag.String("input", "", "file to parse from local disk")
	config := flag.String("config", "", "Substation configuration file")
	timeout := flag.Duration("timeout", 10*time.Second, "timeout")
	transforms := flag.Int("transforms", runtime.NumCPU(), "number of transform goroutines to execute")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	loadConfig(*config)

	if err := file(ctx, *input, *transforms); err != nil {
		panic(err)
	}
}

func file(ctx context.Context, filename string, transforms int) error {
	sub.CreateChannels(transforms)
	defer sub.KillSignal()

	go func() {
		var sinkWg sync.WaitGroup

		sinkWg.Add(1)
		go sub.Sink(ctx, &sinkWg)

		var transformWg sync.WaitGroup
		for w := 1; w <= transforms; w++ {
			transformWg.Add(1)
			go sub.Transform(ctx, &transformWg)
		}

		fileHandle, err := os.Open(filename)
		if err != nil {
			sub.SendErr(err)
			return
		}
		defer fileHandle.Close()
		scanner := bufio.NewScanner(fileHandle)
		for scanner.Scan() {
			sub.SendTransform([]byte(scanner.Text()))
		}

		sub.TransformSignal()
		transformWg.Wait()
		sub.SinkSignal()
		sinkWg.Wait()
	}()

	if err := sub.Block(ctx); err != nil {
		return err
	}

	return nil
}
