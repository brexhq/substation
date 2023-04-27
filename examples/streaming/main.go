// Provides an example of how to use streaming processors, which use channels to
// process data. This example uses the aggregate processor, which aggregates data
// into a single item.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/process"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	conf := config.Config{}
	if err := json.Unmarshal(cfg, &conf); err != nil {
		panic(err)
	}

	// maintains state across goroutines and creates
	// two channels used for ingesting and sinking data
	// from the transform process. Substation channels are
	// unbuffered and can be accessed directly, in addition
	// to helper methods for sending data and closing the
	// channel.
	group, ctx := errgroup.WithContext(context.TODO())
	in, out := config.NewChannel(), config.NewChannel()

	// create and start the transform process. the process
	// must run in a goroutine to prevent blocking
	// the main goroutine.
	agg, err := process.NewStreamer(ctx, conf)
	if err != nil {
		panic(err)
	}

	group.Go(func() error {
		if err := agg.Stream(ctx, in, out); err != nil {
			panic(err)
		}

		return nil
	})

	// reading data must start before writing data and run inside a
	// goroutine to prevent blocking the main goroutine.
	group.Go(func() error {
		for capsule := range out.C {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// do something with the data
				fmt.Printf("sink: %s\n", string(capsule.Data()))
			}
		}

		return nil
	})

	// writing data to the ingest channel can be run from the main goroutine
	// or inside a goroutine. the channel must be closed after all data
	// is written -- this sends a signal to the transform process to
	// stop reading from the channel.
	data := [][]byte{
		[]byte("this"),
		[]byte("is"),
		[]byte("a"),
		[]byte("test"),
	}

	for _, d := range data {
		fmt.Printf("ingest: %s\n", string(d))

		capsule := config.NewCapsule()
		capsule.SetData(d)

		select {
		case <-ctx.Done():
			panic(ctx.Err())
		default:
			in.Send(capsule)
		}
	}
	in.Close()

	// wait for all goroutines to finish, otherwise the program will exit
	// before the transform process completes.
	if err := group.Wait(); err != nil {
		panic(err)
	}
}
