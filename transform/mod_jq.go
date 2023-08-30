package transform

import (
	"context"
	"encoding/json"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	"github.com/brexhq/substation/message"
	"github.com/itchyny/gojq"
)

// errModJQNoOutputGenerated is returned when the jq query generates no output.
var errModJQNoOutputGenerated = fmt.Errorf("no output generated")

type modJQConfig struct {
	// Query is the jq query applied to data.
	Query string `json:"query"`
}

type modJQ struct {
	conf modJQConfig

	query *gojq.Query
}

func newModJQ(_ context.Context, cfg config.Config) (*modJQ, error) {
	conf := modJQConfig{}
	if err := iconfig.Decode(cfg.Settings, &conf); err != nil {
		return nil, fmt.Errorf("transform: new_mod_jq: %v", err)
	}

	// Validate required options.
	if conf.Query == "" {
		return nil, fmt.Errorf("transform: new_mod_jq: query: %v", errors.ErrMissingRequiredOption)
	}

	q, err := gojq.Parse(conf.Query)
	if err != nil {
		return nil, err
	}

	tf := modJQ{
		conf:  conf,
		query: q,
	}

	return &tf, nil
}

func (tf *modJQ) String() string {
	b, _ := gojson.Marshal(tf.conf)
	return string(b)
}

func (*modJQ) Close(context.Context) error {
	return nil
}

func (tf *modJQ) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Skip interrupt messages.
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var i interface{}
	if err := json.Unmarshal(msg.Data(), &i); err != nil {
		return nil, fmt.Errorf("transform: mod_jq: %v", err)
	}

	var arr []interface{}
	iter := tf.query.RunWithContext(ctx, i)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("transform: mod_jq: %v", err)
		}

		arr = append(arr, v)
	}

	var err error
	var b []byte
	switch len(arr) {
	case 0:
		return nil, fmt.Errorf("transform: mod_jq: %v", errModJQNoOutputGenerated)
	case 1:
		b, err = json.Marshal(arr[0])
	default:
		b, err = json.Marshal(arr)
	}

	if err != nil {
		return nil, fmt.Errorf("transform: mod_jq: %v", err)
	}

	outMsg := message.New().SetData(b).SetMetadata(msg.Metadata())
	return []*message.Message{outMsg}, nil
}
