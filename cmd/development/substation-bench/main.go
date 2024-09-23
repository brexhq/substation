// Benchmarks the performance of Substation by sending a configurable number of events
// through the system and reporting the total time taken, the number of events sent, the
// amount of data sent, and the rate of events and data sent per second.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"

	"github.com/brexhq/substation/v2/internal/bufio"
	"github.com/brexhq/substation/v2/internal/channel"
	"github.com/brexhq/substation/v2/internal/file"
)

type options struct {
	Count       int
	Concurrency int
	ConfigFile  string
	DataFile    string
	pprofCPU    bool
	pprofMemory bool
}

func main() {
	var opts options

	flag.StringVar(&opts.DataFile, "file", "", "File to parse")
	flag.IntVar(&opts.Count, "count", 100000, "Number of events to process (default: 100000)")
	flag.IntVar(&opts.Concurrency, "concurrency", -1, "Number of concurrent data transformation functions to run (default: number of CPUs available)")
	flag.StringVar(&opts.ConfigFile, "config", "", "Substation configuration file (default: empty config)")
	flag.BoolVar(&opts.pprofCPU, "cpu", false, "Enable CPU profiling (default: false)")
	flag.BoolVar(&opts.pprofMemory, "mem", false, "Enable memory profiling (default: false)")
	flag.Parse()

	if opts.DataFile == "" {
		fmt.Println("missing required flag -file")
		os.Exit(1)
	}

	ctx := context.Background()

	fmt.Printf("%s: Configuring Substation\n", time.Now().Format(time.RFC3339Nano))
	var conf []byte
	// If no config file is provided, then an empty config is used.
	if opts.ConfigFile != "" {
		path, err := file.Get(ctx, opts.ConfigFile)
		defer os.Remove(path)

		if err != nil {
			panic(err)
		}

		conf, err = os.ReadFile(path)
		if err != nil {
			panic(err)
		}
	} else {
		conf = []byte(`{"transforms":[]}`)
	}

	cfg := substation.Config{}
	if err := json.Unmarshal(conf, &cfg); err != nil {
		panic(err)
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	// Collect the sample data for the benchmark.
	path, err := file.Get(ctx, opts.DataFile)
	defer os.Remove(path)

	if err != nil {
		panic(fmt.Errorf("file: %v", err))
	}

	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("file: %v", err))
	}
	defer f.Close()

	scanner := bufio.NewScanner()
	defer scanner.Close()

	if err := scanner.ReadFile(f); err != nil {
		panic(err)
	}

	fmt.Printf("%s: Loading data into memory\n", time.Now().Format(time.RFC3339Nano))
	var data [][]byte
	dataBytes := 0
	for scanner.Scan() {
		b := []byte(scanner.Text())
		for i := 0; i < opts.Count; i++ {
			data = append(data, b)
			dataBytes += len(b)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
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

	fmt.Printf("%s: Starting benchmark\n", time.Now().Format(time.RFC3339Nano))
	start := time.Now()
	ch := channel.New[*message.Message]()
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		tfGroup, tfCtx := errgroup.WithContext(ctx)
		tfGroup.SetLimit(opts.Concurrency)

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
		if _, err := sub.Transform(ctx, ctrl); err != nil {
			return err
		}

		return nil
	})

	group.Go(func() error {
		defer ch.Close()

		for _, b := range data {
			msg := message.New().SetData(b)
			ch.Send(msg)
		}

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		panic(err)
	}

	fmt.Printf("%s: Ending benchmark\n", time.Now().Format(time.RFC3339Nano))

	// The benchmark reports the total time taken, the number of events sent, the
	// amount of data sent, and the rate of events and data sent per second.
	elapsed := time.Since(start)
	fmt.Printf("\nBenchmark results:\n")
	fmt.Printf("- %d events in %s\n", len(data), elapsed)
	fmt.Printf("- %.2f events per second\n", float64(len(data))/elapsed.Seconds())
	fmt.Printf("- %d MB in %s\n", dataBytes/1000/1000, elapsed)
	fmt.Printf("- %.2f MB per second\n", float64(dataBytes)/1000/1000/elapsed.Seconds())

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
