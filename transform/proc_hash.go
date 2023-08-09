package transform

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	gojson "encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
)

// errProcHashInvalidAlgorithm is returned when the hash transform is configured with an invalid algorithm.
var errProcHashInvalidAlgorithm = fmt.Errorf("invalid algorithm")

type procHashConfig struct {
	// Key retrieves a value from an object for processing.
	//
	// This is optional for transforms that support processing non-object data.
	Key string `json:"key"`
	// SetKey inserts a processed value into an object.
	//
	// This is optional for transforms that support processing non-object data.
	SetKey string `json:"set_key"`
	// Algorithm is the hashing algorithm applied to the data.
	//
	// Must be one of:
	//
	// - md5
	//
	// - sha256
	Algorithm string `json:"algorithm"`
}

type procHash struct {
	conf     procHashConfig
	isObject bool
}

func newProcHash(_ context.Context, cfg config.Config) (*procHash, error) {
	conf := procHashConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if (conf.Key != "" && conf.SetKey == "") ||
		(conf.Key == "" && conf.SetKey != "") {
		return nil, fmt.Errorf("transform: proc_hash: key %s set_key %s: %v", conf.Key, conf.SetKey, errInvalidDataPattern)
	}

	if conf.Algorithm == "" {
		return nil, fmt.Errorf("transform: proc_hash: algorithm: %v", errors.ErrMissingRequiredOption)
	}

	if !slices.Contains(
		[]string{"md5", "sha256"},
		conf.Algorithm) {
		return nil, fmt.Errorf("transform: proc_hash: algorithm %q: %v", conf.Algorithm, errors.ErrInvalidOption)
	}

	proc := procHash{
		conf:     conf,
		isObject: conf.Key != "" && conf.SetKey != "",
	}

	return &proc, nil
}

func (t *procHash) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procHash) Close(context.Context) error {
	return nil
}

func (t *procHash) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		switch t.isObject {
		case true:
			result := message.Get(t.conf.Key).String()

			var value string
			switch t.conf.Algorithm {
			case "md5":
				sum := md5.Sum([]byte(result))
				value = fmt.Sprintf("%x", sum)
			case "sha256":
				sum := sha256.Sum256([]byte(result))
				value = fmt.Sprintf("%x", sum)
			default:
				return nil, fmt.Errorf("transform: proc_hash: algorithm %s: %v", t.conf.Algorithm, errProcHashInvalidAlgorithm)
			}

			if err := message.Set(t.conf.SetKey, value); err != nil {
				return nil, fmt.Errorf("transform: proc_hash: %v", err)
			}

			output = append(output, message)
		case false:
			var value string
			switch t.conf.Algorithm {
			case "md5":
				sum := md5.Sum(message.Data())
				value = fmt.Sprintf("%x", sum)
			case "sha256":
				sum := sha256.Sum256(message.Data())
				value = fmt.Sprintf("%x", sum)
			default:
				return nil, fmt.Errorf("transform: proc_hash: algorithm %s: %v", t.conf.Algorithm, errProcHashInvalidAlgorithm)
			}

			msg, err := mess.New(
				mess.SetData([]byte(value)),
				mess.SetMetadata(message.Metadata()),
			)
			if err != nil {
				return nil, fmt.Errorf("transform: proc_hash: %v", err)
			}

			output = append(output, msg)
		}
	}

	return output, nil
}
