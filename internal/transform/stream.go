package transform

import (
	"context"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/metrics"
	"github.com/brexhq/substation/process"
	"golang.org/x/sync/errgroup"
)

// stream transforms data by applying a series of processors to a pipeline of
// encapsulated data.
//
// Data processing is iterative and each processor is enabled through conditions.
type tformStream struct {
	Processors []config.Config `json:"processors"`

	streamers []process.Streamer
}

func newTformStream(ctx context.Context, cfg config.Config) (t tformStream, err error) {
	if err = config.Decode(cfg.Settings, &t); err != nil {
		return tformStream{}, err
	}

	t.streamers, err = process.NewStreamers(ctx, t.Processors...)
	if err != nil {
		return tformStream{}, err
	}

	return t, nil
}

// Transform processes a channel of encapsulated data with the transform.
func (t tformStream) Transform(ctx context.Context, wg *sync.WaitGroup, in, out *config.Channel) error {
	go func() {
		wg.Wait()
		//nolint: errcheck // errors are ignored in case closing fails in a single applier
		process.CloseStreamers(ctx, t.streamers...)
	}()

	group, ctx := errgroup.WithContext(ctx)

	// each streamer has two channels, one for input and one for output, that create
	// a pipeline. streamers are executed in order as goroutines so that capsules are
	// processed in series.
	prevChan := config.NewChannel()
	firstChan := prevChan

	for _, s := range t.streamers {
		nextChan := config.NewChannel()
		func(s process.Streamer, inner, outer *config.Channel) {
			group.Go(func() error {
				return s.Stream(ctx, inner, outer)
			})
		}(s, prevChan, nextChan)

		prevChan = nextChan
	}

	// the last streamer in the pipeline sends to the sink (drain), and must
	// be read before capsules are put into the pipeline to avoid deadlock.
	var sent int
	group.Go(func() error {
		for capsule := range prevChan.C {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				out.Send(capsule)
				sent++
			}
		}

		return nil
	})

	// the first streamer in the pipeline receives from the source, and must
	// start after the drain goroutine to avoid deadlock.
	var received int
	for capsule := range in.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			firstChan.Send(capsule)
			received++
		}
	}

	// this is required so that the pipeline goroutines can exit.
	firstChan.Close()

	// an error in any streamer will cause the entire pipeline, including sending
	// to the sink, to fail.
	if err := group.Wait(); err != nil {
		return err
	}

	_ = metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesReceived",
		Value: received,
	})

	_ = metrics.Generate(ctx, metrics.Data{
		Name:  "CapsulesSent",
		Value: sent,
	})

	return nil
}
