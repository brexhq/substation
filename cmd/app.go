package cmd

import (
	"context"
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

type cfg struct {
	Transform config.Config
	Sink      config.Config
}

// Substation is the application core, all data processing and flow happens through Substation.
type Substation struct {
	Channels Channels
	Config   cfg
}

/*
Channels contains channels used by the app for managing state and sending encapsulated data between goroutines:

- Done: signals that all data processing (ingest, transform, load) is complete; this is always invoked by the Sink goroutine

- Kill: signals that all non-anonymous goroutines should end processing

- Errs: signals that an error occurred from an internal component

- Transform: sends encapsulated data from the handler to the Transform goroutines

- Sink: sends encapsulated data from the Transform goroutines to the Sink goroutine
*/
type Channels struct {
	Done      chan struct{}
	Kill      chan struct{}
	Errs      chan error
	Transform *config.Channel
	Sink      *config.Channel
	// Transform chan config.Capsule
	// Sink      chan config.Capsule
}

// CreateChannels initializes channels used by the app. Non-blocking channels can leak if the caller closes before processing completes; this is most likely to happen if the caller uses context to timeout. To avoid goroutine leaks, set larger buffer sizes.
func (sub *Substation) CreateChannels(size int) {
	sub.Channels.Done = make(chan struct{})
	sub.Channels.Transform = config.NewChannel()
	sub.Channels.Sink = config.NewChannel()
}

func (sub *Substation) TransformWait(wg *sync.WaitGroup) {
	sub.Channels.Transform.Close()
	wg.Wait()

	log.Debug("closed transform channel")
}

func (sub *Substation) SinkWait(wg *sync.WaitGroup) {
	sub.Channels.Sink.Close()
	wg.Wait()

	log.Debug("closed sink channel")
}

// Send puts byte data into the Transform channel.
func (sub *Substation) Send(cap config.Capsule) {
	sub.Channels.Transform.Send(cap)
}

/*
Block blocks the handler from returning until one of these conditions is met:

- a data processing error occurs

- the request times out (or is otherwise cancelled)

- all data processing is successful

This is usually the final call made by main() in a cmd invoking the app.
*/
func (sub *Substation) Block(ctx context.Context, group *errgroup.Group) error {
	for {
		select {
		case <-ctx.Done():
			// all channels are closed to avoid leaking goroutines
			sub.Channels.Sink.Close()
			sub.Channels.Transform.Close()

			if group.Wait() != nil {
				log.Debug("errored")
				return group.Wait()
			} else {
				log.Debug("cancelled")
				return nil
			}

		case <-sub.Channels.Done:
			log.Debug("finished")
			return nil
		}
	}
}

// Transform is the data transformation method for the app. Data is input on the Transform channel, transformed by a Transform interface (see: internal/transform), and output on the Sink channel. All Transform goroutines complete when the Transform channel is closed and all data is flushed.
func (sub *Substation) Transform(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	t, err := transform.Factory(sub.Config.Transform)
	if err != nil {
		return err
	}

	log.WithField("transform", sub.Config.Transform.Type).Debug("Substation starting transform process")
	if err := t.Transform(ctx, sub.Channels.Transform, sub.Channels.Sink); err != nil {
		return err
	}

	return nil
}

// Sink is the data sink method for the app. Data is input on the Sink channel and sent to the configured sink. The Sink goroutine completes when the Sink channel is closed and all data is flushed.
func (sub *Substation) Sink(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	s, err := sink.Factory(sub.Config.Sink)
	if err != nil {
		return err
	}

	log.WithField("sink", sub.Config.Sink.Type).Debug("Substation starting sink process")
	if err := s.Send(ctx, sub.Channels.Sink); err != nil {
		return err
	}

	close(sub.Channels.Done)

	return nil
}

// GetConcurrency retrieves a concurrency value from the SUBSTATION_CONCURRENCY environment variable. If the environment variable is missing, then the concurrency value is the number of CPUs on the host. In native Substation applications, this value determines the number of transform goroutines; if set to 1, then multi-core processing is not enabled.
func GetConcurrency() (int, error) {
	if val, found := os.LookupEnv("SUBSTATION_CONCURRENCY"); found {
		v, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}
		return v, nil
	}

	return runtime.NumCPU(), nil
}

/*
GetScanMethod retrieves a scan method from the SUBSTATION_SCAN_METHOD environment variable. This impacts the behavior of bufio scanners that are used throughout the application to read files. The options for this variable are:

- "bytes" (https://pkg.go.dev/bufio#Scanner.Bytes)

- "text" (https://pkg.go.dev/bufio#Scanner.Text)

If the environment variable is missing, then the default method is "text".
*/
func GetScanMethod() string {
	if val, found := os.LookupEnv("SUBSTATION_SCAN_METHOD"); found {
		if val == "bytes" || val == "text" {
			return val
		}
	}

	return "text"
}
