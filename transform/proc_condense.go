package transform

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"fmt"
	"sync"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/internal/json"
	mess "github.com/brexhq/substation/message"
	"github.com/jshlbrd/go-aggregate"
)

// errProcCondenseSizeLimit is returned when the condense's buffer size
// limit is reached. If this error occurs, then increase the size of
// the buffer or use the procDrop transform to remove data that exceeds
// the buffer limit.
var errProcCondenseSizeLimit = fmt.Errorf("data exceeded size limit")

type procCondenseConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// CondenseKey retrieves a value from an object that is used to organize
	// condensed objects.
	//
	// This is only used when handling objects and defaults to an
	// empty string.
	CondenseKey string `json:"condense_key"`
	// Separator is the string that joins condensed data.
	//
	// This is only used when handling data and defaults to an empty
	// string.
	Separator string `json:"separator"`
	// MaxCount determines the maximum number of items stored in the
	// buffer before emitting condensed data.
	//
	// This is optional and defaults to 1000 items.
	MaxCount int `json:"max_count"`
	// MaxSize determines the maximum size (in bytes) of items stored
	// in the buffer before emitting condensed data.
	//
	// This is optional and defaults to 10000 (10KB).
	MaxSize int `json:"max_size"`
}

type procCondense struct {
	conf procCondenseConfig

	// buffer is safe for concurrent access.
	mu     sync.Mutex
	buffer map[string]*aggregate.Bytes
}

func newProcCondense(_ context.Context, cfg config.Config) (*procCondense, error) {
	conf := procCondenseConfig{}
	if err := config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	if conf.MaxCount == 0 {
		conf.MaxCount = 1000
	}

	if conf.MaxSize == 0 {
		conf.MaxSize = 10000
	}

	proc := procCondense{
		conf: conf,
	}

	proc.mu = sync.Mutex{}
	proc.buffer = make(map[string]*aggregate.Bytes)

	return &proc, nil
}

func (t *procCondense) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procCondense) Close(context.Context) error {
	return nil
}

func (t *procCondense) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
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
			return nil, fmt.Errorf("transform: proc_condense: size %d data length %d: %v", t.conf.MaxSize, length, errProcCondenseSizeLimit)
		}

		var condenseKey string
		if t.conf.CondenseKey != "" {
			condenseKey = message.Get(t.conf.CondenseKey).String()
		}

		if _, ok := t.buffer[condenseKey]; !ok {
			t.buffer[condenseKey] = &aggregate.Bytes{}
			t.buffer[condenseKey].New(t.conf.MaxCount, t.conf.MaxSize)
		}

		ok := t.buffer[condenseKey].Add(message.Data())
		// Data was successfully added to the buffer, every item after
		// this is a failure.
		if ok {
			continue
		}

		data := t.buffer[condenseKey].Get()
		c, err := t.newMessage(data)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_condense: %v", err)
		}
		output = append(output, c)

		// By this point, addition of the failed data is guaranteed
		// to succeed after the buffer is reset.
		t.buffer[condenseKey].Reset()
		_ = t.buffer[condenseKey].Add(message.Data())
	}

	// Drains items from the buffer. If a control was received, then
	// data is emitted regardless of the buffer limits. Otherwise,
	// data is emitted when the buffer limits are reached.
	for key := range t.buffer {
		fmt.Println("k:", key)

		if control {
			goto CTRL
		}

		if t.buffer[key].Count() < t.conf.MaxCount && t.buffer[key].Size() < t.conf.MaxSize {
			continue
		}

	CTRL:
		data := t.buffer[key].Get()
		msg, err := t.newMessage(data)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_condense: %v", err)
		}

		output = append(output, msg)
		delete(t.buffer, key)
	}

	return output, nil
}

func (t *procCondense) newMessage(data [][]byte) (*mess.Message, error) {
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
