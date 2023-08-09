package transform

import (
	"context"
	"encoding/json"
	gojson "encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	_config "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/errors"
	mess "github.com/brexhq/substation/message"
	"github.com/itchyny/gojq"
)

// errProcJQNoOutputGenerated is returned when the jq query generates no output.
var errProcJQNoOutputGenerated = fmt.Errorf("no output generated")

type procJQConfig struct {
	// Query is the jq query applied to data.
	Query string `json:"query"`
}

type procJQ struct {
	conf procJQConfig

	query *gojq.Query
}

func newProcJQ(_ context.Context, cfg config.Config) (*procJQ, error) {
	conf := procJQConfig{}
	if err := _config.Decode(cfg.Settings, &conf); err != nil {
		return nil, err
	}

	// Validate required options.
	if conf.Query == "" {
		return nil, fmt.Errorf("transform: proc_jq: query: %v", errors.ErrMissingRequiredOption)
	}

	q, err := gojq.Parse(conf.Query)
	if err != nil {
		return nil, err
	}

	proc := procJQ{
		conf:  conf,
		query: q,
	}

	return &proc, nil
}

func (t *procJQ) String() string {
	b, _ := gojson.Marshal(t.conf)
	return string(b)
}

func (*procJQ) Close(context.Context) error {
	return nil
}

func (t *procJQ) Transform(ctx context.Context, messages ...*mess.Message) ([]*mess.Message, error) {
	var output []*mess.Message

	for _, message := range messages {
		// Skip control messages.
		if message.IsControl() {
			output = append(output, message)
			continue
		}

		var i interface{}
		if err := json.Unmarshal(message.Data(), &i); err != nil {
			return nil, fmt.Errorf("transform: proc_jq: %v", err)
		}

		var arr []interface{}
		iter := t.query.RunWithContext(ctx, i)

		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return nil, fmt.Errorf("transform: proc_jq: %v", err)
			}

			arr = append(arr, v)
		}

		var err error
		var b []byte
		switch len(arr) {
		case 0:
			err = errProcJQNoOutputGenerated
		case 1:
			b, err = json.Marshal(arr[0])
			if err != nil {
				return nil, fmt.Errorf("transform: proc_jq: %v", err)
			}
		default:
			b, err = json.Marshal(arr)
		}

		if err != nil {
			return nil, fmt.Errorf("transform: proc_jq: %v", err)
		}

		msg, err := mess.New(
			mess.SetData(b),
			mess.SetMetadata(message.Metadata()),
		)
		if err != nil {
			return nil, fmt.Errorf("transform: proc_jq: %v", err)
		}

		output = append(output, msg)
	}

	return output, nil
}
