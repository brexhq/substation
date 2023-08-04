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

	"github.com/brexhq/substation"
	"github.com/brexhq/substation/internal/bufio"
	"github.com/brexhq/substation/internal/channel"
	"github.com/brexhq/substation/internal/file"
	mess "github.com/brexhq/substation/message"
	"github.com/brexhq/substation/transform"
	"golang.org/x/sync/errgroup"
)

type options struct {
	Count       int
	Parallelism int
	ConfigFile  string
	DataFile    string
	pprofCPU    bool
	pprofMemory bool
}

// nolint: gocognit // Ignore cognitive complexity.
func main() {
	var opts options

	flag.StringVar(&opts.DataFile, "input", "", "path to sample data file")
	flag.IntVar(&opts.Count, "count", 100000, "number of events to send")
	flag.IntVar(&opts.Parallelism, "parallelism", 1, "number of data transformation goroutines")
	flag.StringVar(&opts.ConfigFile, "config", "", "path to configuration file (optional)")
	flag.BoolVar(&opts.pprofCPU, "cpu", false, "enable CPU profiling (optional)")
	flag.BoolVar(&opts.pprofMemory, "mem", false, "enable memory profiling (optional)")
	flag.Parse()

	if opts.DataFile == "" {
		fmt.Println("missing required flag -input")
		os.Exit(1)
	}

	ctx := context.Background()

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
		conf = []byte(`{"transform":[]}`)
	}

	cfg := substation.Config{}
	if err := json.Unmarshal(conf, &cfg); err != nil {
		panic(err)
	}

	sub, err := substation.New(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer sub.Close(ctx)

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

	var data []byte
	for scanner.Scan() {
		switch scanner.Method() {
		case "bytes":
			data = scanner.Bytes()
		case "text":
			data = []byte(scanner.Text())
		}

		// Only read the first line of the file.
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

	start := time.Now()
	ch := channel.New[*mess.Message]()
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		group, ctx := errgroup.WithContext(ctx)
		group.SetLimit(opts.Parallelism)

		for message := range ch.Recv() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			message := message
			group.Go(func() error {
				if _, err := transform.Apply(ctx, sub.Transforms(), message); err != nil {
					return err
				}

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			return err
		}

		return nil
	})

	group.Go(func() error {
		defer ch.Close()

		for i := 0; i < opts.Count; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			message, err := mess.New(mess.SetData(data))
			if err != nil {
				return err
			}

			ch.Send(message)
		}

		control, err := mess.New(mess.AsControl())
		if err != nil {
			return err
		}
		ch.Send(control)

		return nil
	})

	// Wait for all goroutines to complete. This includes the goroutines that are
	// executing the transform functions.
	if err := group.Wait(); err != nil {
		panic(err)
	}

	// The benchmark reports the total time taken, the number of events sent, the
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
