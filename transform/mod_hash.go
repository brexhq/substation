package transform

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
)

type modHashConfig struct {
	Object configObject `json:"object"`

	// Algorithm is the hashing algorithm applied to the data.
	//
	// Must be one of:
	//
	// - MD5
	//
	// - SHA256
	Algorithm string `json:"algorithm"`
}

type modHash struct {
	conf     modHashConfig
	isObject bool
}

func newModHash(_ context.Context, cfg config.Config) (*modHash, error) {
	conf := modHashConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_hash: %v", err)
	}

	// Validate required options.
	if conf.Object.Key == "" && conf.Object.SetKey != "" {
		return nil, fmt.Errorf("transform: new_mod_hash: object_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Object.Key != "" && conf.Object.SetKey == "" {
		return nil, fmt.Errorf("transform: new_mod_hash: object_set_key: %v", errors.ErrMissingRequiredOption)
	}

	if conf.Algorithm == "" {
		return nil, fmt.Errorf("transform: new_mod_hash: algorithm: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{"md5", "MD5", "sha256", "SHA256"},
		conf.Algorithm) {
		return nil, fmt.Errorf("transform: new_mod_hash: algorithm %q: %v", conf.Algorithm, errors.ErrInvalidOption)
	}

	tf := modHash{
		conf:     conf,
		isObject: conf.Object.Key != "" && conf.Object.SetKey != "",
	}

	return &tf, nil
}

func (tf *modHash) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modHash) Close(context.Context) error {
	return nil
}

func (tf *modHash) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	if !tf.isObject {
		var value string
		switch tf.conf.Algorithm {
		case "md5", "MD5":
			sum := md5.Sum(msg.Data())
			value = fmt.Sprintf("%x", sum)
		case "sha256", "SHA256":
			sum := sha256.Sum256(msg.Data())
			value = fmt.Sprintf("%x", sum)
		}

		data := []byte(value)
		finMsg := message.New().SetData(data).SetMetadata(msg.Metadata())
		return []*message.Message{finMsg}, nil
	}

	result := msg.GetObject(tf.conf.Object.Key).String()

	var value string
	switch tf.conf.Algorithm {
	case "md5", "MD5":
		sum := md5.Sum([]byte(result))
		value = fmt.Sprintf("%x", sum)
	case "sha256", "SHA256":
		sum := sha256.Sum256([]byte(result))
		value = fmt.Sprintf("%x", sum)
	}

	if err := msg.SetObject(tf.conf.Object.SetKey, value); err != nil {
		return nil, fmt.Errorf("transform: mod_hash: %v", err)
	}

	return []*message.Message{msg}, nil
}
