// Benchmarks the performance of Substation by sending a configurable number of events
// through the system and reporting the total time taken, the number of events sent, the
// amount of data sent, and the rate of events and data sent per second.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/brexhq/substation/cmd"
	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/file"
	"golang.org/x/sync/errgroup"
)

type options struct {
	Count       int
	ConfigFile  string
	DataFile    string
	pprofCPU    bool
	pprofMemory bool
}

func main() {
	var opts options

	flag.StringVar(&opts.DataFile, "input", "", "path to sample data file")
	flag.IntVar(&opts.Count, "count", 100000, "number of events to send")
	flag.StringVar(&opts.ConfigFile, "config", "", "path to configuration file (optional)")
	flag.BoolVar(&opts.pprofCPU, "cpu", false, "enable CPU profiling (optional)")
	flag.BoolVar(&opts.pprofMemory, "mem", false, "enable memory profiling (optional)")
	flag.Parse()

	if opts.DataFile == "" {
		fmt.Println("missing required flag -input")
		os.Exit(1)
	}

	ctx := context.Background()

	var cfg []byte

	// if no config file is provided, then a noop config (data transfer only)
	// is used. benchmarking external services is possible by configuring a
	// sink, otherwise the config should use the noop sink.
	if opts.ConfigFile != "" {
		path, err := file.Get(ctx, opts.ConfigFile)
		defer os.Remove(path)

		if err != nil {
			panic(err)
		}

		cfg, err = os.ReadFile(path)
		if err != nil {
			panic(err)
		}
	} else {
		cfg = []byte(`{"sink":{"type":"noop"},"transform":{"type":"noop"}}`)
	}

	sub := cmd.New()
	if err := sub.SetConfig(bytes.NewReader(cfg)); err != nil {
		panic(err)
	}

	// collect the sample data for the benchmark
	path, err := file.Get(ctx, opts.DataFile)
	defer os.Remove(path)

	if err != nil {
		panic(err)
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner()
	defer scanner.Close()

	if err := scanner.ReadFile(f); err != nil {
		panic(err)
	}

	var data []byte
	for scanner.Scan() {
		switch scanner.Method() {
		case "bytes":
			data = scanner.Bytes()
		case "text":
			data = []byte(scanner.Text())
		}

		// only read the first line of the file
		//nolint:staticcheck // ignore SA4004
		break
	}

	if opts.pprofCPU {
		f, err := os.Create("./cpu.prof")
		if err != nil {
			panic(err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	// this benchmarks a simulated application that includes:
	// - ingesting data
	// - transforming data
	// - loading data
	start := time.Now()
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

	group.Go(func() error {
		for i := 0; i < opts.Count; i++ {
			capsule := config.NewCapsule()
			capsule.SetData(data)

			sub.Send(capsule)
		}

		sub.WaitTransform(&transformWg)
		sub.WaitSink(&sinkWg)

		return nil
	})

	if err := sub.Block(ctx, group); err != nil {
		panic(err)
	}

	// the benchmark reports the total time taken, the number of events sent, the
	// amount of data sent, and the rate of events and data sent per second.
	elapsed := time.Since(start)
	fmt.Printf("%d events in %s (%.2f events/sec)\n", opts.Count, elapsed, float64(opts.Count)/elapsed.Seconds())
	fmt.Printf("%d MB in %s (%.2f MB/sec)\n", opts.Count*len(data)/1024/1024, elapsed, float64(opts.Count*len(data))/1024/1024/elapsed.Seconds())

	if opts.pprofMemory {
		heap, err := os.Create("./heap.prof")
		if err != nil {
			panic(err)
		}
		if err := pprof.WriteHeapProfile(heap); err != nil {
			panic(err)
		}
	}
}
