package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"

	"github.com/brexhq/substation/v2/config"
	"github.com/brexhq/substation/v2/message"

	iconfig "github.com/brexhq/substation/v2/internal/config"
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
		return fmt.Errorf("query: %v", iconfig.ErrMissingRequiredOption)
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

	// UseNumber keeps integers exact: gojq turns a json.Number into an int,
	// or a *big.Int when it does not fit, instead of rounding to float64.
	dec := json.NewDecoder(bytes.NewReader(msg.Data()))
	dec.UseNumber()

	var i interface{}
	if err := dec.Decode(&i); err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// Reject trailing data after the value, as json.Unmarshal does. The
	// trailing token is left out of the error so payload data is not logged.
	if _, err := dec.Token(); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
		}
		return nil, fmt.Errorf("transform %s: unexpected data after top-level value", tf.conf.ID)
	}

	i, err := objectJQFixNumbers(i)
	if err != nil {
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

// Preserve json.Unmarshal behavior that gojq's number normalization changes.
func objectJQFixNumbers(v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		for k, e := range v {
			n, err := objectJQFixNumbers(e)
			if err != nil {
				return nil, err
			}
			v[k] = n
		}
	case []interface{}:
		for k, e := range v {
			n, err := objectJQFixNumbers(e)
			if err != nil {
				return nil, err
			}
			v[k] = n
		}
	case json.Number:
		s := v.String()
		if s == "-0" {
			return math.Copysign(0, -1), nil
		}
		if strings.ContainsAny(s, ".eE") {
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return nil, err
			}
		}
	}

	return v, nil
}

func (tf *objectJQ) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
