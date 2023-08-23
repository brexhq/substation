package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/aggregate"
	_config "github.com/brexhq/substation/internal/config"
	mess "github.com/brexhq/substation/message"
)

type procCombineConfig struct {
	Buffer aggregate.Config `json:"buffer"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// CombineKey retrieves a value from an object that is used to organize
	// combined objects.
	//
	// This is only used when handling objects and defaults to an
	// empty string.
	CombineKey string `json:"combine_key"`
	// Separator is the string that joins combined data.
	//
	// This is only used when handling data and defaults to an empty
	// string.
	Separator string `json:"separator"`
}

type procCombine struct {
	conf procCombineConfig

	// buffer is safe for concurrent access.
	mu        sync.Mutex
	buffer    map[string]*aggregate.Aggregate
	bufferCfg aggregate.Config
}

func newProcCombine(_ context.Context, cfg config.Config) (*procCombine, error) {
	conf := procCombineConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	proc := procCombine{
		conf: conf,
	}

	proc.mu = sync.Mutex{}
	proc.buffer = make(map[string]*aggregate.Aggregate)
	proc.bufferCfg = aggregate.Config{
		Count:    conf.Buffer.Count,
		Size:     conf.Buffer.Size,
		Interval: conf.Buffer.Interval,
	}

	return &proc, nil
}

func (proc *procCombine) String() string {
	b, _ := gojson.Marshal(proc.conf)
	return string(b)
}

func (*procCombine) Close(context.Context) error {
	return nil
}

func (proc *procCombine) Transform(ctx context.Context, message *mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	proc.mu.Lock()
	defer proc.mu.Unlock()

	if message.IsControl() {
		var output []*mess.Message

		for key := range proc.buffer {
			msg, err := proc.newMessage(proc.buffer[key].Get())
			if err != nil {
				return nil, fmt.Errorf("transform: proc_combine: %v", err)
			}

			output = append(output, msg)
		}

		proc.buffer = make(map[string]*aggregate.Aggregate)
		output = append(output, message)

		return output, nil
	}

	var combineKey string
	if proc.conf.CombineKey != "" {
		combineKey = message.Get(proc.conf.CombineKey).String()
	}

	if _, ok := proc.buffer[combineKey]; !ok {
		agg, err := aggregate.New(proc.bufferCfg)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_combine: %v", err)
		}

		proc.buffer[combineKey] = agg
	}

	if ok := proc.buffer[combineKey].Add(message.Data()); ok {
		return nil, nil
	}

	msg, err := proc.newMessage(proc.buffer[combineKey].Get())
	if err != nil {
		return nil, fmt.Errorf("transform: proc_combine: %v", err)
	}

	// By this point, addition of the failed data is guaranteed
	// to succeed after the buffer is reset.
	proc.buffer[combineKey].Reset()
	_ = proc.buffer[combineKey].Add(message.Data())

	return []*mess.Message{msg}, nil
}

func (proc *procCombine) newMessage(data [][]byte) (*mess.Message, error) {
	if proc.conf.SetKey != "" {
		msg, err := mess.New()
		if err != nil {
			return nil, err
		}

		for _, d := range data {
			if err := msg.Set(proc.conf.SetKey+".-1", d); err != nil {
				return nil, err
			}
		}

		return msg, nil
	}

	value := bytes.Join(data, []byte(proc.conf.Separator))
	return mess.New(
		mess.SetData(value),
	)
}
