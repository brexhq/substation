package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/brexhq/substation/internal/base64"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	// metaKey is a prefix used to access the meta field in a Message.
	metaKey = "meta "
)

var (
	// errSetRawInvalidValue is returned when setRaw receives an invalid interface type.
	errSetRawInvalidValue = fmt.Errorf("invalid value type")

	sjsonOpts = &sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	}
)

// Message is the data structure that is handled by transforms and interpreted by
// conditions.
//
// Data in each message may be JSON text or binary data:
//   - JSON text is accessed using the GetValue, SetValue, and DeleteValue methods.
//   - Binary data is accessed using the Data and SetData methods.
//
// Metadata is a second data field that is meant to store information about the message,
// but can be used for any purpose. For JSON text, metadata is accessed using the
// GetValue, SetValue, and DeleteValue methods with a key prefixed with "meta ". Binary
// metadata is accessed using the Metadata and SetMetadata methods.
//
// Messages can also be configured as "control messages." Control messages are used for flow
// control in Substation functions and applications, but can be used for any purpose depending
// on the needs of a transform or condition. These messages should not contain data or metadata.
type Message struct {
	data []byte
	meta []byte

	// ctrl is a flag that indicates if the message is a control message.
	//
	// Control messages trigger special behavior in transforms and conditions.
	ctrl bool
}

func (m *Message) String() string {
	return string(m.data)
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

func (m *Message) Data() []byte {
	if m.ctrl {
		return nil
	}

	return m.data
}

func (m *Message) SetData(data []byte) *Message {
	if m.ctrl {
		return m
	}

	m.data = data
	return m
}

func (m *Message) Metadata() []byte {
	if m.ctrl {
		return nil
	}

	return m.meta
}

func (m *Message) SetMetadata(metadata []byte) *Message {
	if m.ctrl {
		return m
	}

	m.meta = metadata
	return m
}

func (m *Message) GetValue(key string) Value {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)

		v := gjson.GetBytes(m.meta, key)
		return Value{gjson: v}
	}

	key = strings.TrimSpace(key)
	v := gjson.GetBytes(m.data, key)
	return Value{gjson: v}
}

func (m *Message) SetValue(key string, value interface{}) error {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)

		meta, err := setValue(m.meta, key, value)
		if err != nil {
			return err
		}
		m.meta = meta

		return nil
	}

	key = strings.TrimSpace(key)
	data, err := setValue(m.data, key, value)
	if err != nil {
		return err
	}
	m.data = data

	return nil
}

func (m *Message) DeleteValue(key string) error {
	if strings.HasPrefix(key, metaKey) {
		key = strings.TrimPrefix(key, metaKey)
		key = strings.TrimSpace(key)

		meta, err := deleteValue(m.meta, key)
		if err != nil {
			return err
		}
		m.meta = meta

		return nil
	}

	data, err := deleteValue(m.data, key)
	if err != nil {
		return err
	}
	m.data = data

	return nil
}

// Value is a wrapper around gjson.Result that provides a consistent interface
// for converting values from JSON text.
type Value struct {
	gjson gjson.Result
}

func (v Value) Value() interface{} {
	return v.gjson.Value()
}

func (v Value) String() string {
	return v.gjson.String()
}

func (v Value) Bytes() []byte {
	return []byte(v.gjson.String())
}

func (v Value) Int() int64 {
	return v.gjson.Int()
}

func (v Value) Uint() uint64 {
	return v.gjson.Uint()
}

func (v Value) Float() float64 {
	return v.gjson.Float()
}

func (v Value) Bool() bool {
	return v.gjson.Bool()
}

func (v Value) Array() []Value {
	var values []Value
	for _, r := range v.gjson.Array() {
		values = append(values, Value{gjson: r})
	}

	return values
}

func (v Value) IsArray() bool {
	return v.gjson.IsArray()
}

func (v Value) Map() map[string]Value {
	values := make(map[string]Value)
	for k, r := range v.gjson.Map() {
		values[k] = Value{gjson: r}
	}

	return values
}

func (v Value) Exists() bool {
	return v.gjson.Exists()
}

func deleteValue(json []byte, key string) ([]byte, error) {
	b, err := sjson.DeleteBytes(json, key)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func setValue(json []byte, key string, value interface{}) (b []byte, err error) {
	if validJSON(value) {
		return setRaw(json, key, value)
	}

	switch v := value.(type) {
	case []byte:
		if utf8.Valid(v) {
			b, err = sjson.SetBytesOptions(json, key, v, sjsonOpts)
		} else {
			b, err = sjson.SetBytesOptions(json, key, base64.Encode(v), sjsonOpts)
		}
	case Value:
		b, err = sjson.SetBytesOptions(json, key, v.Value(), sjsonOpts)
	default:
		b, err = sjson.SetBytesOptions(json, key, v, sjsonOpts)
	}

	if err != nil {
		return nil, err
	}

	return b, nil
}

func setRaw(json []byte, key string, value interface{}) (b []byte, err error) {
	switch v := value.(type) {
	case []byte:
		b, err = sjson.SetRawBytesOptions(json, key, v, sjsonOpts)
	case string:
		b, err = sjson.SetRawBytesOptions(json, key, []byte(v), sjsonOpts)
	case Value:
		b, err = sjson.SetRawBytesOptions(json, key, v.Bytes(), sjsonOpts)
	default:
		return nil, errSetRawInvalidValue
	}

	if err != nil {
		return nil, err
	}

	return b, nil
}

func validJSON(data interface{}) bool {
	switch v := data.(type) {
	case []byte:
		if !bytes.HasPrefix(v, []byte(`{`)) && !bytes.HasPrefix(v, []byte(`[`)) {
			return false
		}

		return json.Valid(v)
	case string:
		if !strings.HasPrefix(v, `{`) && !strings.HasPrefix(v, `[`) {
			return false
		}

		return json.Valid([]byte(v))
	case Value:
		s := v.String()
		if !strings.HasPrefix(s, `{`) && !strings.HasPrefix(s, `[`) {
			return false
		}

		return json.Valid([]byte(s))
	default:
		return false
	}
}
