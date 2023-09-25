package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/itchyny/gojq"
)

// errObjectJQNoOutputGenerated is returned when the jq query generates no output.
var errObjectJQNoOutputGenerated = fmt.Errorf("no output generated")

type objectJQConfig struct {
	// Query is the jq query applied to data.
	Query string `json:"query"`
}

func (c *objectJQConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectJQConfig) Validate() error {
	if c.Query == "" {
		return fmt.Errorf("query: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectJQ(_ context.Context, cfg config.Config) (*objectJQ, error) {
	conf := objectJQConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: new_object_jq: %v", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform: new_object_jq: %v", err)
	}

	q, err := gojq.Parse(conf.Query)
	if err != nil {
		return nil, err
	}

	tf := objectJQ{
		conf:  conf,
		query: q,
	}

	return &tf, nil
}

type objectJQ struct {
	conf objectJQConfig

	query *gojq.Query
}

func (tf *objectJQ) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var i interface{}
	if err := json.Unmarshal(msg.Data(), &i); err != nil {
		return nil, fmt.Errorf("transform: object_jq: %v", err)
	}

	var arr []interface{}
	iter := tf.query.RunWithContext(ctx, i)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("transform: object_jq: %v", err)
		}

		arr = append(arr, v)
	}

	var err error
	var b []byte
	switch len(arr) {
	case 0:
		return nil, fmt.Errorf("transform: object_jq: %v", errObjectJQNoOutputGenerated)
	case 1:
		b, err = json.Marshal(arr[0])
	default:
		b, err = json.Marshal(arr)
	}

	if err != nil {
		return nil, fmt.Errorf("transform: object_jq: %v", err)
	}

	msg.SetData(b)
	return []*message.Message{msg}, nil
}

func (tf *objectJQ) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
