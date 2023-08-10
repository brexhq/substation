package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
	"github.com/jshlbrd/go-aggregate"
)

// errProcCombineSizeLimit is returned when the buffer size limit is reached.
// If this error occurs, then increase the size of the buffer or use the procDrop
// transform to remove data that exceeds the buffer limit.
var errProcCombineSizeLimit = fmt.Errorf("data exceeded size limit")

type procCombineConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
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
	// MaxCount determines the maximum number of items stored in the
	// buffer before emitting combined data.
	//
	// This is optional and defaults to 1000 items.
	MaxCount int `json:"max_count"`
	// MaxSize determines the maximum size (in bytes) of items stored
	// in the buffer before emitting combined data.
	//
	// This is optional and defaults to 10000 (10KB).
	MaxSize int `json:"max_size"`
}

type procCombine struct {
	conf procCombineConfig

	// buffer is safe for concurrent access.
	mu     sync.Mutex
	buffer map[string]*aggregate.Bytes
}

func newProcCombine(_ context.Context, cfg config.Config) (*procCombine, error) {
	conf := procCombineConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	if conf.MaxCount == 0 {
		conf.MaxCount = 1000
	}

	if conf.MaxSize == 0 {
		conf.MaxSize = 10000
	}

	proc := procCombine{
		conf: conf,
	}

	proc.mu = sync.Mutex{}
	proc.buffer = make(map[string]*aggregate.Bytes)

	return &proc, nil
}

func (t *procCombine) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procCombine) Close(context.Context) error {
	return nil
}

func (t *procCombine) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	// Lock the transform to prevent concurrent access to the buffer.
	t.mu.Lock()
	defer t.mu.Unlock()

	var control bool
	var output []*mess.Message

	for _, message := range messages {
		if message.IsControl() {
			control = true
			output = append(output, message)
			continue
		}

		// Data that exceeds the size of the buffer will never
		// fit within it.
		length := len(message.Data())
		if length > t.conf.MaxSize {
			return nil, fmt.Errorf("transform: proc_combine: size %d data length %d: %v", t.conf.MaxSize, length, errProcCombineSizeLimit)
		}

		var combineKey string
		if t.conf.CombineKey != "" {
			combineKey = message.Get(t.conf.CombineKey).String()
		}

		if _, ok := t.buffer[combineKey]; !ok {
			t.buffer[combineKey] = &aggregate.Bytes{}
			t.buffer[combineKey].New(t.conf.MaxCount, t.conf.MaxSize)
		}

		ok := t.buffer[combineKey].Add(message.Data())
		// Data was successfully added to the buffer, every item after
		// this is a failure.
		if ok {
			continue
		}

		data := t.buffer[combineKey].Get()
		c, err := t.newMessage(data)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_combine: %v", err)
		}
		output = append(output, c)

		// By this point, addition of the failed data is guaranteed
		// to succeed after the buffer is reset.
		t.buffer[combineKey].Reset()
		_ = t.buffer[combineKey].Add(message.Data())
	}

	// If a control message was received, then items are flushed from the buffer.
	if !control {
		return messages, nil
	}

	for key := range t.buffer {
		data := t.buffer[key].Get()
		msg, err := t.newMessage(data)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_combine: %v", err)
		}

		output = append(output, msg)
		delete(t.buffer, key)
	}

	return output, nil
}

func (t *procCombine) newMessage(data [][]byte) (*mess.Message, error) {
	if t.conf.SetKey != "" {
		var value []byte
		for _, element := range data {
			var err error

			value, err = json.Set(value, t.conf.SetKey, element)
			if err != nil {
				return nil, err
			}
		}

		return mess.New(
			mess.SetData(value),
		)
	}

	value := bytes.Join(data, []byte(t.conf.Separator))
	return mess.New(
		mess.SetData(value),
	)
}
