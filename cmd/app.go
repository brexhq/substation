// package cmd provides definitions and methods for building Substation applications.
package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/log"
	"github.com/brexhq/substation/internal/sink"
	"github.com/brexhq/substation/internal/transform"
	"golang.org/x/sync/errgroup"
)

// substation is the application core that manages all data processing and flow control.
type substation struct {
	config      cfg
	channels    channels
	concurrency int
}

// cfg is the shared application configuration for all apps.
type cfg struct {
	Transform config.Config
	Sink      config.Config
}

/*
channels contains channels used by the app for managing state and sending encapsulated data between goroutines:

- done: signals that all data processing (ingest, transform, load) is complete; this is always invoked by the Sink goroutine

- transform: sends encapsulated data from the source application to the Transform goroutines

- sink: sends encapsulated data from the Transform goroutines to the Sink goroutine
*/
type channels struct {
	done      chan struct{}
	transform *config.Channel
	sink      *config.Channel
}

/*
New returns an initialized Substation app. If an error occurs during initialization, then this function will panic.

Concurrency is controlled using the SUBSTATION_CONCURRENCY environment variable and defaults to the number of CPUs on the host. In native Substation applications, this value determines the number of transform goroutines; if set to 1, then multi-core processing is not enabled.
*/
func New() *substation {
	sub := &substation{}

	sub.config = cfg{}
	sub.channels.done = make(chan struct{})
	sub.channels.transform = config.NewChannel()
	sub.channels.sink = config.NewChannel()

	sub.concurrency = runtime.NumCPU()
	val, found := os.LookupEnv("SUBSTATION_CONCURRENCY")
	if found {
		v, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		sub.concurrency = v
	}

	return sub
}

// SetConfig loads a configuration into the app.
func (sub *substation) SetConfig(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(&sub.config); err != nil {
		return err
	}

	return nil
}

// Concurrency returns the concurrency setting of the app.
func (sub *substation) Concurrency() int {
	return sub.concurrency
}

// SetConcurrency sets the concurrency setting of the app. This method overrides the default concurrency that is set when the app is created.
func (sub *substation) SetConcurrency(c int) {
	sub.concurrency = c
}

// Send writes encapsulated data into the Transform channel.
func (sub *substation) Send(capsule config.Capsule) {
	sub.channels.transform.Send(capsule)
}

/*
Block blocks the handler from returning until one of these conditions is met:

- a data processing error occurs

- the request times out (or is otherwise cancelled)

- all data processing is successful

This is usually the final call made by main() in a cmd invoking the app.
*/
func (sub *substation) Block(ctx context.Context, group *errgroup.Group) error {
	for {
		select {
		// ctx must be derived from the group using WithContext and
		// carries error and cancellation signals for all goroutines
		case <-ctx.Done():
			// all channels are closed to address an edge case where
			// a producer goroutine hangs when putting an item into a
			// channel where the consumer goroutine has terminated
			//
			// this mitigates unintentional freezing of the source
			// application and leaking its goroutines
			sub.channels.sink.Close()
			sub.channels.transform.Close()

			if group.Wait() != nil {
				log.Debug("processing errored")
				return group.Wait()
			} else {
				log.Debug("processing cancelled")
				return ctx.Err()
			}

		// signals that all data processing completed successfully
		// this should only ever be called by Sink
		case <-sub.channels.done:
			log.Debug("processing finished")
			return nil
		}
	}
}

// Transform is the data transformation method for the app. Data is input on the Transform channel, transformed by a Transform interface (see: internal/transform), and output on the Sink channel. All Transform goroutines complete when the Transform channel is closed and all data is flushed.
func (sub *substation) Transform(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	t, err := transform.Factory(sub.config.Transform)
	if err != nil {
		return err
	}

	log.WithField("transform", sub.config.Transform.Type).Debug("starting transformer")
	if err := t.Transform(ctx, wg, sub.channels.transform, sub.channels.sink); err != nil {
		return err
	}

	return nil
}

// WaitTransform closes the transform channel and blocks until data processing is complete.
func (sub *substation) WaitTransform(wg *sync.WaitGroup) {
	sub.channels.transform.Close()
	wg.Wait()

	log.Debug("transformers finished")
}

// Sink is the data sink method for the app. Data is input on the Sink channel and sent to the configured sink. The Sink goroutine completes when the Sink channel is closed and all data is flushed.
func (sub *substation) Sink(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	s, err := sink.Factory(sub.config.Sink)
	if err != nil {
		return err
	}

	log.WithField("sink", sub.config.Sink.Type).Debug("starting sink")
	if err := s.Send(ctx, sub.channels.sink); err != nil {
		return err
	}

	close(sub.channels.done)

	return nil
}

// WaitSink closes the sink channel and blocks until data load is complete.
func (sub *substation) WaitSink(wg *sync.WaitGroup) {
	sub.channels.sink.Close()
	wg.Wait()

	log.Debug("sink finished")
}
