package message

import (
	"fmt"
	"strings"

	"github.com/brexhq/substation/internal/json"
	"github.com/tidwall/gjson"
)

const (
	// metaKey is a prefix used to access the meta field in a Message.
	metaKey = "meta "
)

var errControlWithData = fmt.Errorf("control message cannot contain data")

func New(options ...func(*Message)) (*Message, error) {
	m := &Message{}
	for _, o := range options {
		o(m)
	}

	// Control messages cannot contain data.
	if m.ctrl && (len(m.data) > 0 || len(m.meta) > 0) {
		return m, errControlWithData
	}

	return m, nil
}

func SetData(data []byte) func(*Message) {
	return func(m *Message) {
		m.data = data
	}
}

func SetMetadata(metadata []byte) func(*Message) {
	return func(m *Message) {
		m.meta = metadata
	}
}

func AsControl() func(*Message) {
	return func(m *Message) {
		m.ctrl = true
	}
}

type Message struct {
	// If data is an object, then it is accessed using the Get, Set,
	// and Delete methods. The field can have its value returned
	// directly by using the Data method.
	data []byte

	// If metadata is an object, then it is accessed using the Get, Set,
	// and Delete methods along with the "metadata" key prefix (e.g.,
	// "metadata [key]"). The field can have its value returned directly
	// by using the Metadata method.
	meta []byte

	// ctrl is a flag that indicates if the message is a control message.
	//
	// Control messages trigger special behavior in data transforms.
	// For example, a control message may be used by a transform to emit
	// data that was previously buffered.
	//
	// Control messages cannot contain data or metadata.
	ctrl bool
}

// Delete removes a key from objects stored in the Message.
func (m *Message) Delete(key string) (err error) {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)

		if key == "@this" {
			m.meta = nil
			return nil
		}

		meta, err := json.Delete(m.meta, key)
		if err != nil {
			return err
		}
		m.meta = meta

		return nil
	}

	if key == "@this" {
		m.data = nil
		return nil
	}

	data, err := json.Delete(m.data, key)
	if err != nil {
		return err
	}
	m.data = data

	return nil
}

// Get retrieves a value from objects in the Message. Metadata is accessed using the
// "metadata" key prefix: "metadata [key]". If the key is empty, then no value is
// returned.
func (m *Message) Get(key string) gjson.Result {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)
		return json.Get(m.meta, key)
	}

	key = strings.TrimSpace(key)
	return json.Get(m.data, key)
}

// Set writes a value to objects in the Message.
func (m *Message) Set(key string, value interface{}) (err error) {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)

		meta, err := json.Set(m.meta, key, value)
		if err != nil {
			return err
		}
		m.meta = meta

		return nil
	}

	key = strings.TrimSpace(key)
	data, err := json.Set(m.data, key, value)
	if err != nil {
		return err
	}
	m.data = data

	return nil
}

func (m *Message) Data() []byte {
	if m.ctrl {
		return nil
	}

	return m.data
}

func (m *Message) Metadata() []byte {
	if m.ctrl {
		return nil
	}

	return m.meta
}

func (m *Message) IsControl() bool {
	return m.ctrl
}
