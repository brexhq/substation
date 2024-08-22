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

// errObjectJQNoOutputGenerated is returned when jq generates no output.
var errObjectJQNoOutputGenerated = fmt.Errorf("no output generated")

type objectJQConfig struct {
	// Filter is the jq filter applied to data.
	Filter string `json:"filter"`

	ID string `json:"id"`
}

func (c *objectJQConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func (c *objectJQConfig) Validate() error {
	if c.Filter == "" {
		return fmt.Errorf("query: %v", errors.ErrMissingRequiredOption)
	}

	return nil
}

func newObjectJQ(_ context.Context, cfg config.Config) (*objectJQ, error) {
	conf := objectJQConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform object_jq: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "object_jq"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	q, err := gojq.Parse(conf.Filter)
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
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	var arr []interface{}
	iter := tf.query.RunWithContext(ctx, i)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}

		arr = append(arr, v)
	}

	var err error
	var b []byte
	switch len(arr) {
	case 0:
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, errObjectJQNoOutputGenerated)
	case 1:
		b, err = json.Marshal(arr[0])
	default:
		b, err = json.Marshal(arr)
	}

	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	msg.SetData(b)
	return []*message.Message{msg}, nil
}

func (tf *objectJQ) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
