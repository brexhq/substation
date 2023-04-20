// package config provides capabilities for managing configurations and handling data in Substation applications.
package config

import (
	gojson "encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/brexhq/substation/internal/json"
)

// errSetInvalidKey is returned when an invalid key is used in a Capsule Set function.
var errSetInvalidKey = fmt.Errorf("invalid set key")

// Config is a template used by Substation interface factories to produce new instances from JSON configurations. Type refers to the type of instance and Settings contains options used in the instance. Examples of this are found in the condition and process packages.
type Config struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

// Decode marshals and unmarshals an input interface into the output interface using the standard library's json package. This should be used when decoding JSON configurations (i.e., Config) in Substation interface factories.
func Decode(input, output interface{}) error {
	b, err := gojson.Marshal(input)
	if err != nil {
		return err
	}
	return gojson.Unmarshal(b, output)
}

// Get retrieves a configuration location from the SUBSTATION_CONFIG environment variable.
func Get() string {
	if v, found := os.LookupEnv("SUBSTATION_CONFIG"); found {
		return v
	}

	return ""
}

/*
Capsule stores encapsulated data that is used throughout the package's data handling and processing functions.

Each capsule contains two unexported fields that are accessed by getters and setters:

- data: stores structured or unstructured data

- metadata: stores structured metadata that describes the data

Values in the metadata field are accessed using the pattern "!metadata [key]". JSON values can be freely moved between the data and metadata fields.

Substation applications follow these rules when handling capsules:

- Sources set the initial metadata, but this can be modified in transit by applying processors

- Sinks only output data, but metadata can be retained by copying it from metadata into data
*/
type Capsule struct {
	data     []byte
	metadata []byte
}

// NewCapsule returns a new, empty Capsule.
func NewCapsule() Capsule {
	return Capsule{}
}

// Delete removes a key from a JSON object stored in the capsule's data or metadata fields.
func (c *Capsule) Delete(key string) (err error) {
	if strings.HasPrefix(key, "!metadata") {
		key = strings.TrimPrefix(key, "!metadata")
		key = strings.TrimLeft(key, " ")

		if key == "" {
			c.metadata = nil
			return nil
		}

		c.metadata, err = json.Delete(c.metadata, key)
		if err != nil {
			return err
		}

		return nil
	}

	c.data, err = json.Delete(c.data, key)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves a value from a JSON object stored in the capsule's data or metadata fields.
func (c *Capsule) Get(key string) json.Result {
	if strings.HasPrefix(key, "!metadata") {
		key = strings.TrimPrefix(key, "!metadata")
		key = strings.TrimLeft(key, " ")

		// returns entire metadata object
		if key == "" {
			return json.Get(c.metadata, "@this")
		}

		return json.Get(c.metadata, key)
	}

	return json.Get(c.data, key)
}

// Set writes a value to a JSON object stored in the capsule's data or metadata fields.
func (c *Capsule) Set(key string, value interface{}) (err error) {
	if strings.HasPrefix(key, "!metadata") {
		key = strings.TrimPrefix(key, "!metadata")
		key = strings.TrimLeft(key, " ")

		// values should not be written directly to the metadata field
		if key == "" {
			return errSetInvalidKey
		}

		c.metadata, err = json.Set(c.metadata, key, value)
		if err != nil {
			return err
		}

		return nil
	}

	c.data, err = json.Set(c.data, key, value)
	if err != nil {
		return err
	}

	return nil
}

// SetRaw writes a raw value to a JSON object stored in the capsule's data or metadata fields. These values are usually pre-formatted JSON (e.g., entire objects or arrays).
func (c *Capsule) SetRaw(key string, value interface{}) (err error) {
	if strings.HasPrefix(key, "!metadata ") {
		key = strings.TrimPrefix(key, "!metadata ")

		// values should not be written directly to the metadata field
		if key == "" {
			return errSetInvalidKey
		}

		c.metadata, err = json.SetRaw(c.metadata, key, value)
		if err != nil {
			return err
		}

		return nil
	}

	c.data, err = json.SetRaw(c.data, key, value)
	if err != nil {
		return err
	}

	return nil
}

// Data returns the contents of the capsule's data field.
func (c *Capsule) Data() []byte {
	return c.data
}

// Metadata returns the contents of the capsule's metadata field.
func (c *Capsule) Metadata() []byte {
	return c.metadata
}

// SetData writes data to the capsule's data field.
func (c *Capsule) SetData(b []byte) *Capsule {
	c.data = b
	return c
}

// SetMetadata writes data to the capsule's metadata field. Metadata must be an interface that can marshal to a JSON object.
func (c *Capsule) SetMetadata(i interface{}) (*Capsule, error) {
	meta, err := gojson.Marshal(i)
	if err != nil {
		return nil, err
	}

	c.metadata = meta
	return c, nil
}

// Channel provides methods for safely writing capsule data to and closing channels. Data should be read directly from the channel (e.g., ch.C).
type Channel struct {
	C      chan Capsule
	mu     sync.Mutex
	closed bool
}

// NewChannel returns an unbuffered channel.
func NewChannel() *Channel {
	return &Channel{C: make(chan Capsule)}
}

// Close closes a channel and relies on a mutex to prevent panicking if the channel is closed by multiple goroutines.
func (c *Channel) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		close(c.C)
		c.closed = true
	}
}

// Send writes capsule data to a channel and relies on goroutine recovery to prevent panicking if writes are attempted on a closed channel. Read more about recovery here: https://go.dev/blog/defer-panic-and-recover.
func (c *Channel) Send(capsule Capsule) {
	defer func() {
		//nolint: errcheck // we expect this recover to happen and have no error handling
		// to do.
		recover()
	}()

	c.C <- capsule
}
