package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
)

func init() {
	rootCmd.AddCommand(demoCmd)
}

const demoConf = `
local sub = import '../../substation.libsonnet';

{
  transforms: [
    // Move the event to the 'event.original' field.
    sub.tf.obj.cp({object: { source_key: '@this', target_key: 'meta event.original' }}),
    sub.tf.obj.cp({object: { source_key: 'meta @this' }}),

    // Insert the hash of the original event into the 'event.hash' field.
    sub.tf.hash.sha256({obj: { src: 'event.original', trg: 'event.hash'}}),

    // Insert the event dataset into the 'event.dataset' field.
    sub.tf.obj.insert({obj: { trg: 'event.dataset' }, value: 'aws.cloudtrail'}),

    // Insert the kind of event into the 'event.kind' field.
    sub.tf.obj.insert({obj: { trg: 'event.kind' }, value: 'event'}),

    // Insert the event category into the 'event.category' field.
    sub.tf.obj.insert({obj: { trg: std.format('%s.-1', 'event.category') }, value: 'configuration'}),

    // Insert the event type into the 'event.type' field.
    sub.tf.obj.insert({obj: { trg: std.format('%s.-1', 'event.type') }, value: 'change'}),

    // Insert the outcome into the 'event.outcome' field.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.num.len.gt({ obj: { src: 'errorCode' }, value: 0 }),
        transforms: [
          sub.tf.obj.insert({ obj: { trg: 'event.outcome' }, value: 'failure' }),
        ],
      },
      {
        transforms: [
          sub.tf.obj.insert({ obj: { trg: 'event.outcome' }, value: 'success' }),
        ],
      },
    ] }),

    // Copy the event time to the '@timestamp' field.
    sub.tf.obj.cp({obj: { src: 'event.original.eventTime', trg: '\\@timestamp' }}),

    // Copy the IP address to the 'source.ip' field.
    sub.tf.obj.cp({obj: { src: 'event.original.sourceIPAddress', trg: 'source.ip' }}),

    // Copy the user agent to the 'user_agent.original' field.
    sub.tf.obj.cp({obj: { src: 'event.original.userAgent', trg: 'user_agent.original' }}),

    // Copy the region to the 'cloud.region' field.
    sub.tf.obj.cp({obj: { src: 'event.original.awsRegion', trg: 'cloud.region' }}),

    // Copy the account ID to the 'cloud.account.id' field.
    sub.tf.obj.cp({obj: { src: 'event.original.userIdentity.accountId', trg: 'cloud.account.id' }}),

    // Add the cloud service provider to the 'cloud.provider' field.
    sub.tf.obj.insert({obj: { trg: 'cloud.provider' }, value: 'aws'}),

    // Extract the cloud service into the 'cloud.service.name' field.
    sub.tf.str.capture({obj: { src: 'event.original.eventSource', trg: 'cloud.service.name' }, pattern: '^(.*)\\.amazonaws\\.com$'}),

    // Make the event pretty before printing to the console.
    sub.tf.obj.cp({obj: { src: '@this|@pretty' }}),
    sub.tf.send.stdout(),
  ],
}		
`

const demoCompiled = `
{
   "transforms": [
      {
         "settings": {
            "id": "2bbe3748-28c56e0b",
            "object": {
               "source_key": "@this",
               "target_key": "meta event.original"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "2bbe3748-61e51827",
            "object": {
               "source_key": "meta @this"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "324f1035-f49e5682",
            "object": {
               "source_key": "event.original",
               "target_key": "event.hash"
            }
         },
         "type": "hash_sha256"
      },
      {
         "settings": {
            "id": "5f4ae672-0478e109",
            "object": {
               "target_key": "event.dataset"
            },
            "value": "aws.cloudtrail"
         },
         "type": "object_insert"
      },
      {
         "settings": {
            "id": "5f4ae672-7de9f731",
            "object": {
               "target_key": "event.kind"
            },
            "value": "event"
         },
         "type": "object_insert"
      },
      {
         "settings": {
            "id": "5f4ae672-2c1fa54f",
            "object": {
               "target_key": "event.category.-1"
            },
            "value": "configuration"
         },
         "type": "object_insert"
      },
      {
         "settings": {
            "id": "5f4ae672-e97ed8b8",
            "object": {
               "target_key": "event.type.-1"
            },
            "value": "change"
         },
         "type": "object_insert"
      },
      {
         "settings": {
            "cases": [
               {
                  "condition": {
                     "settings": {
                        "measurement": "byte",
                        "object": {
                           "source_key": "errorCode"
                        },
                        "value": 0
                     },
                     "type": "number_length_greater_than"
                  },
                  "transforms": [
                     {
                        "settings": {
                           "id": "5f4ae672-c3cc893e",
                           "object": {
                              "target_key": "event.outcome"
                           },
                           "value": "failure"
                        },
                        "type": "object_insert"
                     }
                  ]
               },
               {
                  "transforms": [
                     {
                        "settings": {
                           "id": "5f4ae672-87ff6d17",
                           "object": {
                              "target_key": "event.outcome"
                           },
                           "value": "success"
                        },
                        "type": "object_insert"
                     }
                  ]
               }
            ],
            "id": "b3a47dd1-fddb5674"
         },
         "type": "meta_switch"
      },
      {
         "settings": {
            "id": "2bbe3748-e3640864",
            "object": {
               "source_key": "event.original.eventTime",
               "target_key": "\\@timestamp"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "2bbe3748-63faf2a6",
            "object": {
               "source_key": "event.original.sourceIPAddress",
               "target_key": "source.ip"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "2bbe3748-3b7dfda5",
            "object": {
               "source_key": "event.original.userAgent",
               "target_key": "user_agent.original"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "2bbe3748-626bded4",
            "object": {
               "source_key": "event.original.awsRegion",
               "target_key": "cloud.region"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "2bbe3748-061dfac7",
            "object": {
               "source_key": "event.original.userIdentity.accountId",
               "target_key": "cloud.account.id"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "id": "5f4ae672-5c9e5d3a",
            "object": {
               "target_key": "cloud.provider"
            },
            "value": "aws"
         },
         "type": "object_insert"
      },
      {
         "settings": {
            "count": 0,
            "id": "e3bd5484-53bd3692",
            "object": {
               "source_key": "event.original.eventSource",
               "target_key": "cloud.service.name"
            },
            "pattern": "^(.*)\\.amazonaws\\.com$"
         },
         "type": "string_capture"
      },
      {
         "settings": {
            "id": "2bbe3748-15552062",
            "object": {
               "source_key": "@this|@pretty"
            }
         },
         "type": "object_copy"
      },
      {
         "settings": {
            "batch": {
               "count": 1000,
               "duration": "1m",
               "size": 1000000
            },
            "id": "de19b3c9-67c1890d"
         },
         "type": "send_stdout"
      }
   ]
}
`

const demoEvt = `{"eventVersion":"1.08","userIdentity":{"type":"IAMUser","principalId":"EXAMPLE123456789","arn":"arn:aws:iam::123456789012:user/Alice","accountId":"123456789012","accessKeyId":"ASIAEXAMPLE123","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2024-10-01T12:00:00Z"},"sessionIssuer":{"type":"AWS","principalId":"EXAMPLE123456","arn":"arn:aws:iam::123456789012:role/Admin","accountId":"123456789012","userName":"Admin"}}},"eventTime":"2024-10-01T12:30:45Z","eventSource":"s3.amazonaws.com","eventName":"PutBucketPolicy","awsRegion":"us-west-2","sourceIPAddress":"203.0.113.0","userAgent":"aws-sdk-python/1.0.0 Python/3.8.0 Linux/4.15.0","requestParameters":{"bucketName":"example-bucket","policy":"{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"s3:GetObject\",\"Resource\":\"arn:aws:s3:::example-bucket/*\"}]}"}},"responseElements":{"location":"http://example-bucket.s3.amazonaws.com/"},"requestID":"EXAMPLE123456789","eventID":"EXAMPLE-1-2-3-4-5-6","readOnly":false,"resources":[{"ARN":"arn:aws:s3:::example-bucket","accountId":"123456789012","type":"AWS::S3::Bucket"}],"eventType":"AwsApiCall","managementEvent":true,"recipientAccountId":"123456789012"}`

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "demo substation",
	Long: `'substation demo' shows how Substation transforms data.
It prints an anonymized CloudTrail event (input) and the
transformed result (output) to the console. The event is 
partially normalized to the Elastic Common Schema (ECS).
`,
	// Examples:
	//  substation demo
	Example: `  substation demo
`,
	Args: cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := substation.Config{}

		if err := json.Unmarshal([]byte(demoCompiled), &cfg); err != nil {
			return err
		}

		ctx := context.Background() // This doesn't need to be canceled.
		sub, err := substation.New(ctx, cfg)
		if err != nil {
			return err
		}

		msgs := []*message.Message{
			message.New().SetData([]byte(demoEvt)),
			message.New().AsControl(),
		}

		// Make the input pretty before printing to the console.
		fmt.Printf("input:\n%s\n", gjson.Get(demoEvt, "@this|@pretty").String())
		fmt.Printf("output:\n")

		if _, err := sub.Transform(ctx, msgs...); err != nil {
			return err
		}

		fmt.Printf("\nconfig:%s\n", demoConf)

		return nil
	},
}
