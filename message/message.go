package message

import (
	"strings"

	"github.com/brexhq/substation/internal/json"
	"github.com/tidwall/gjson"
)

const (
	// metaKey is a prefix used to access the meta field in a Message.
	metaKey = "meta "
)

// Message is the data structure that is passed between transforms and
// interpretable by conditions.
type Message struct {
	// If data is JSON text, then it is accessed using the GetJSON, SetJSON,
	// and DeleteJSON methods. The field can have its value returned
	// directly by using the Data method.
	data []byte

	// If meta is JSON text, then it is accessed using the GetJSON, SetJSON,
	// and DeleteJSON methods using the "meta" key prefix (e.g., "meta [key]").
	// The field can have its value returned directly by using the Meta method.
	meta []byte

	// ctrl is a flag that indicates if the message is a control message.
	//
	// Control messages trigger special behavior in data transforms.
	// For example, they can be used to force a transform to emit buffered data.
	//
	// These messages should not contain data or metadata.
	ctrl bool
}

func New(opts ...func(*Message)) *Message {
	msg := &Message{}
	for _, o := range opts {
		o(msg)
	}

	return msg
}

func AsControl() func(*Message) {
	return func(m *Message) {
		m.ctrl = true
	}
}

func (m *Message) IsControl() bool {
	return m.ctrl
}

func (m *Message) GetObject(key string) gjson.Result {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)
		return json.Get(m.meta, key)
	}

	key = strings.TrimSpace(key)
	return json.Get(m.data, key)
}

func (m *Message) SetObject(key string, value interface{}) error {
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

func (m *Message) DeleteObject(key string) error {
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

func (m *Message) SetData(data []byte) *Message {
	if m.ctrl {
		return m
	}

	m.data = data
	return m
}

func (m *Message) Data() []byte {
	if m.ctrl {
		return nil
	}

	return m.data
}

func (m *Message) SetMetadata(metadata []byte) *Message {
	if m.ctrl {
		return m
	}

	m.meta = metadata
	return m
}

func (m *Message) Metadata() []byte {
	if m.ctrl {
		return nil
	}

	return m.meta
}
