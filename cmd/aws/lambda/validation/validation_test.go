package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/brexhq/substation/internal/errors"
)

var testCfgs = []struct {
	name        string
	cfg         []byte
	expectedErr error
}{
	{
		"invalid transform",
		[]byte(`
		{
			"transforms": [
				{
				   "type": "fooer"
				}
			]
		 }
		 `),
		errors.ErrInvalidFactoryInput,
	},
	{
		"invalid processor settings",
		[]byte(`
		{
			"transforms": [
				{
					"type": "hash",
					"settings": {
						"options": {
							"algorithm": "md1"
						}
					}
				}
			]
		}					
		 `),
		errors.ErrInvalidOption,
	},
	{
		"valid config",
		[]byte(`
		{
			"transforms": [
				{
					"settings": {
					   "key": "foo",
					   "set_key": "baz"
					},
					"type": "proc_copy"
				 }
			]
		 }
		 `),
		nil,
	},
}

func TestHandler(t *testing.T) {
	for _, cfg := range testCfgs {
		t.Run(cfg.name, func(t *testing.T) {
			e, err := json.Marshal(validationEvent{
				Content: base64.StdEncoding.EncodeToString(cfg.cfg),
				URI:     "arn:aws:lambda:region:account:function:SubstationAppConfigLambdaValidator",
			})
			if err != nil {
				t.Fatal(err)
			}

			err = handler(context.Background(), e)
			if err != nil && cfg.expectedErr == nil {
				t.Error(err)
			}
		})
	}
}
