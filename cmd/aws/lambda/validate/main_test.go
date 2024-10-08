package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	iconfig "github.com/brexhq/substation/v2/internal/config"
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
		iconfig.ErrInvalidFactoryInput,
	},
	{
		"invalid processor settings",
		[]byte(`
		{
			"transforms": [
				{
					"type": "object_insert",
				}
			]
		}					
		 `),
		iconfig.ErrInvalidOption,
	},
	{
		"valid config",
		[]byte(`
		{
			"transforms": [
				{
					"settings": {
						"object": {
							"source_key": "foo",
							"target_key": "baz"	 
						}
					},
					"type": "object_copy"
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
