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
		"invalid sink",
		[]byte(`
		{
			"sink": {
			   "type": "fooer"
			},
			"transform": {
			   "type": "transfer"
			}
		 }
		 `),
		errors.ErrInvalidFactoryInput,
	},
	{
		"invalid transform",
		[]byte(`
		{
			"sink": {
			   "type": "stdout"
			},
			"transform": {
			   "type": "fooer"
			}
		 }
		 `),
		errors.ErrInvalidFactoryInput,
	},
	{
		"invalid processor",
		[]byte(`
		{
			"sink": {
				"type": "stdout"
			},
			"transform": {
				"settings": {
				   "processors": [
					  {
						 "type": "fooer"
					  }
				   ]
				},
				"type": "batch"
			 }
		 }
		 `),
		errors.ErrInvalidFactoryInput,
	},
	{
		"invalid processor settings",
		[]byte(`
		{
			"sink": {
			  "type": "stdout"
			},
			"transform": {
			  "settings": {
				"processors": [
				  {
					"settings": {
					  "options": {
						"algorithm": "md1"
					  }
					},
					"type": "hash"
				  }
				]
			  },
			  "type": "batch"
			}
		  }
		 `),
		errors.ErrInvalidOption,
	},
	{
		"valid config",
		[]byte(`
		{
			"sink": {
				"type": "stdout"
			},
			"transform": {
				"settings": {
				   "processors": [
					{
						"settings": {
						   "input_key": "foo",
						   "output_key": "baz"
						},
						"type": "copy"
					 }		 
				   ]
				},
				"type": "batch"
			 }
		 }
		 `),
		nil,
	},
}

func TestHandler(t *testing.T) {
	for _, cfg := range testCfgs {
		t.Run(cfg.name, func(t *testing.T) {
			e, err := json.Marshal(validationEvent{
				Content: base64.RawStdEncoding.EncodeToString(cfg.cfg),
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
