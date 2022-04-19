package cmd

import (
	"context"
	"sync"

	"github.com/brexhq/substation/internal/log"
	"github.com/brexhq/substation/internal/sink"
	"github.com/brexhq/substation/internal/transform"
)

type config struct {
	Transform transform.Config
	Sink      sink.Config
}

// Substation is the application core, all data processing and flow happens through Substation.
type Substation struct {
	Channels Channels
	Config   config
}

/*
Channels contains channels used by the app for managing state and transferring data.
	Done: used for signaling that all data processing (ingest, transform, load) is complete; this is always invoked by the Sink goroutine
	Kill: used for signaling that all non-anonymous goroutines should end processing
	Errs: used for signaling that an error occurred from an internal component
	Transform: used for transferring data from the handler to the Transform goroutines
	Sink: used for passing data from the Transform goroutines to the Sink goroutine
*/
type Channels struct {
	Done      chan struct{}
	Kill      chan struct{}
	Errs      chan error
	Transform chan []byte
	Sink      chan []byte
}

// CreateChannels initializes channels used by the app. Non-blocking channels can leak if the caller closes before processing completes; this is most likely to happen if the caller uses context to timeout. To avoid goroutine leaks, set larger buffer sizes.
func (sub *Substation) CreateChannels(size int) {
	sub.Channels.Done = make(chan struct{})
	sub.Channels.Kill = make(chan struct{})
	sub.Channels.Errs = make(chan error, size)
	sub.Channels.Transform = make(chan []byte, size)
	sub.Channels.Sink = make(chan []byte, size)
}

// DoneSignal closes the Done channel. This signals that all data was sent to a sink. This should only be called by the Sink goroutine.
func (sub *Substation) DoneSignal() {
	log.Debug("Substation done signal received, closing done channel")
	close(sub.Channels.Done)
}

// KillSignal closes the Kill channel. This signals all non-anonymous goroutines to stop running. This should always be deferred by the cmd invoking the app.
func (sub *Substation) KillSignal() {
	log.Debug("Substation kill signal received, closing kill channel")
	close(sub.Channels.Kill)
}

// TransformSignal closes the Transform channel. This signals that there is no more incoming data to process. This should only be called by the cmd invoking the app.
func (sub *Substation) TransformSignal() {
	log.Debug("Substation transform signal received, closing transform channel")
	close(sub.Channels.Transform)
}

// SinkSignal closes the Sink channel. This signals that there is no more data to send. This should only be called by the cmd invoking the app.
func (sub *Substation) SinkSignal() {
	log.Debug("Substation sink signal received, closing sink channel")
	close(sub.Channels.Sink)
}

// SendTransform puts byte data into the Transform channel.
func (sub *Substation) SendTransform(b []byte) {
	sub.Channels.Transform <- b
}

// SendErr puts an error into the Errs channel.
func (sub *Substation) SendErr(err error) {
	sub.Channels.Errs <- err
}

/*
Block blocks the handler from returning until one of these conditions is met:
- the handler request times out (ctx.Done)
- a data processing error occurs
- all data processing is complete

This is usually the final call made by main() in a cmd invoking the app.
*/
func (sub *Substation) Block(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.WithField("err", ctx.Err()).Debug("Substation received context signal")
			return ctx.Err()
		case err := <-sub.Channels.Errs:
			log.WithField("err", err).Debug("Substation received error signal")
			return err
		case <-sub.Channels.Done:
			log.Debug("Substation received done signal")
			return nil
		}
	}
}

// Transform is the data transformation method for the app. Data is input on the Transform channel, transformed by a Transform interface (see: internal/transform), and output on the Sink channel. All Transform goroutines finish when the Transform channel is closed.
func (sub *Substation) Transform(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	t, err := transform.Factory(sub.Config.Transform)
	if err != nil {
		sub.SendErr(err)
		return
	}

	log.WithField("transform", sub.Config.Transform.Type).Debug("Substation starting transform process")
	if err := t.Transform(ctx, sub.Channels.Transform, sub.Channels.Sink, sub.Channels.Kill); err != nil {
		sub.SendErr(err)
		return
	}
}

// Sink is the data sink method for the app. Data is input on the Sink channel and sent to the configured sink. Sink finished when the Sink channel is closed.
func (sub *Substation) Sink(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	s, err := sink.Factory(sub.Config.Sink)
	if err != nil {
		sub.SendErr(err)
		return
	}

	log.WithField("sink", sub.Config.Sink.Type).Debug("Substation starting sink process")
	if err := s.Send(ctx, sub.Channels.Sink, sub.Channels.Kill); err != nil {
		sub.SendErr(err)
		return
	}

	sub.DoneSignal()
}
