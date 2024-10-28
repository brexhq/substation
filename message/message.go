// Package message provides functions for managing data used by conditions and transforms.
package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/brexhq/substation/v2/internal/base64"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Flag int

const (
	// metaKey is a prefix used to access the meta field in a Message.
	metaKey = "meta "

	// IsControl indicates that the message is a control message.
	IsControl Flag = iota
	// SkipNullValues indicates that null values should be ignored when processing the message.
	SkipNullValues
	// SkipMissingValues indicates that missing values should be ignored when processing the message.
	SkipMissingValues
	// SkipEmptyValues indicates that empty values should be ignored when processing the message.
	SkipEmptyValues
)

// errSetRawInvalidValue is returned when setRaw receives an invalid interface type.
var errSetRawInvalidValue = fmt.Errorf("invalid value type")

// Message is the data structure that is handled by transforms and interpreted by
// conditions.
//
// Data in each message can be accessed and modified as JSON text or binary data:
//   - JSON text is accessed using the GetValue, SetValue, and DeleteValue methods.
//   - Binary data is accessed using the Data and SetData methods.
//
// Metadata is an additional data field that is meant to store information about the
// message, but can be used for any purpose. For JSON text, metadata is accessed using
// the GetValue, SetValue, and DeleteValue methods with a key prefixed with "meta" (e.g.
// "meta foo"). Binary metadata is accessed using the Metadata and SetMetadata methods.
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

	flags []Flag
}

// String returns the message data as a string.
func (m *Message) String() string {
	return string(m.data)
}

// New returns a new Message.
func New(opts ...func(*Message)) *Message {
	msg := &Message{}
	for _, o := range opts {
		o(msg)
	}

	return msg
}

// HasFlag returns true if the message contains a flag.
func (m *Message) HasFlag(i Flag) bool {
	for _, f := range m.flags {
		if f == i {
			return true
		}
	}

	return false
}

func (m *Message) SkipNullValues() *Message {
	m.flags = append(m.flags, SkipNullValues)
	return m
}

func (m *Message) SkipMissingValues() *Message {
	m.flags = append(m.flags, SkipMissingValues)
	return m
}

func (m *Message) SkipEmptyValues() *Message {
	m.flags = append(m.flags, SkipEmptyValues)
	return m
}

// AsControl sets the message as a control message.
func (m *Message) AsControl() *Message {
	m.data = nil
	m.meta = nil

	m.ctrl = true
	m.flags = append(m.flags, IsControl)

	return m
}

// IsControl returns true if the message is a control message.
func (m *Message) IsControl() bool {
	return m.ctrl
}

// Data returns the message data.
func (m *Message) Data() []byte {
	if m.ctrl {
		return nil
	}

	return m.data
}

// SetData sets the message data.
func (m *Message) SetData(data []byte) *Message {
	if m.ctrl {
		return m
	}

	m.data = data
	return m
}

// Metadata returns the message metadata.
func (m *Message) Metadata() []byte {
	if m.ctrl {
		return nil
	}

	return m.meta
}

// SetMetadata sets the message metadata.
func (m *Message) SetMetadata(metadata []byte) *Message {
	if m.ctrl {
		return m
	}

	m.meta = metadata
	return m
}

// GetValue returns a value from the message data or metadata.
//
// If the key is prefixed with "meta" (e.g. "meta foo"), then
// the value is retrieved from the metadata field, otherwise it
// is retrieved from the data field.
//
// This only works with JSON text. If the message data or metadata
// is not JSON text, then an empty value is returned.
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

// SetValue sets a value in the message data or metadata.
//
// If the key is prefixed with "meta" (e.g. "meta foo"), then
// the value is placed into the metadata field, otherwise it
// is placed into the data field.
//
// This only works with JSON text. If the message data or metadata
// is not JSON text, then this method does nothing.
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

// DeleteValue deletes a value in the message data or metadata.
//
// If the key is prefixed with "meta" (e.g. "meta foo"), then
// the value is removed from the metadata field, otherwise it
// is removed from the data field.
//
// This only works with JSON text. If the message data or metadata
// is not JSON text, then this method does nothing.
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

// Value returns the value as an interface{}.
func (v Value) Value() any {
	return v.gjson.Value()
}

// String returns the value as a string.
func (v Value) String() string {
	return v.gjson.String()
}

// Bytes returns the value as a byte slice.
func (v Value) Bytes() []byte {
	return []byte(v.gjson.String())
}

// Int returns the value as an int64.
func (v Value) Int() int64 {
	return v.gjson.Int()
}

// Uint returns the value as a uint64.
func (v Value) Uint() uint64 {
	return v.gjson.Uint()
}

// Float returns the value as a float64.
func (v Value) Float() float64 {
	return v.gjson.Float()
}

// Bool returns the value as a bool.
func (v Value) Bool() bool {
	return v.gjson.Bool()
}

// Array returns the value as a slice of Value.
func (v Value) Array() []Value {
	var values []Value
	for _, r := range v.gjson.Array() {
		values = append(values, Value{gjson: r})
	}

	return values
}

// IsArray returns true if the value is an array.
func (v Value) IsArray() bool {
	return v.gjson.IsArray()
}

// Map returns the value as a map of string to Value.
func (v Value) Map() map[string]Value {
	values := make(map[string]Value)
	for k, r := range v.gjson.Map() {
		values[k] = Value{gjson: r}
	}

	return values
}

// IsObject returns true if the value is an object.
func (v Value) IsObject() bool {
	return v.gjson.IsObject()
}

// Exists returns true if the value exists.
func (v Value) Exists() bool {
	return v.gjson.Exists()
}

// IsNull returns true if the value is null.
func (v Value) IsNull() bool {
	return v.gjson.Type == gjson.Null
}

// IsMissing returns true if the value is missing.
func (v Value) IsMissing() bool {
	return !v.gjson.Exists()
}

// IsEmpty returns true if the value is an empty string,
// empty array, empty object, or null.
func (v Value) IsEmpty() bool {
	if v.IsArray() {
		return len(v.Array()) == 0
	}

	if v.IsObject() {
		return v.String() == "{}"
	}

	// This catches all other types, including strings and null.
	return v.String() == ""
}

func deleteValue(json []byte, key string) ([]byte, error) {
	b, err := sjson.DeleteBytes(json, key)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// sjson.SetBytesOptions is not used because transform benchmarks perform better with
// sjson.SetBytes (allocating a new byte slice). This may change if transforms are
// refactored.
func setValue(json []byte, key string, value interface{}) ([]byte, error) {
	if validJSON(value) {
		return setRaw(json, key, value)
	}

	switch v := value.(type) {
	case []byte:
		if utf8.Valid(v) {
			return sjson.SetBytes(json, key, v)
		} else {
			return sjson.SetBytes(json, key, base64.Encode(v))
		}
	case Value:
		return sjson.SetBytes(json, key, v.Value())
	default:
		return sjson.SetBytes(json, key, v)
	}
}

// sjson.SetRawBytesOptions is not used because transform benchmarks perform better with
// sjson.SetRawBytes (allocating a new byte slice). This may change if transforms are
// refactored.
func setRaw(json []byte, key string, value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return sjson.SetRawBytes(json, key, v)
	case string:
		return sjson.SetRawBytes(json, key, []byte(v))
	case Value:
		return sjson.SetRawBytes(json, key, v.Bytes())
	default:
		return nil, errSetRawInvalidValue
	}
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
		return validJSON(v.String())
	default:
		return false
	}
}
