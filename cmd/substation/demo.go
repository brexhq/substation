package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/v2"
	"github.com/brexhq/substation/v2/message"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
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
		res, err := vm.EvaluateAnonymousSnippet("demo", demoConf)
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(res), &cfg); err != nil {
			return err
		}

		ctx := context.Background() // This doesn't need to be canceled.
		sub, err := substation.New(ctx, cfg)
		if err != nil {
			return err
		}

		evt := `{"eventVersion":"1.08","userIdentity":{"type":"IAMUser","principalId":"EXAMPLE123456789","arn":"arn:aws:iam::123456789012:user/Alice","accountId":"123456789012","accessKeyId":"ASIAEXAMPLE123","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2024-10-01T12:00:00Z"},"sessionIssuer":{"type":"AWS","principalId":"EXAMPLE123456","arn":"arn:aws:iam::123456789012:role/Admin","accountId":"123456789012","userName":"Admin"}}},"eventTime":"2024-10-01T12:30:45Z","eventSource":"s3.amazonaws.com","eventName":"PutBucketPolicy","awsRegion":"us-west-2","sourceIPAddress":"203.0.113.0","userAgent":"aws-sdk-python/1.0.0 Python/3.8.0 Linux/4.15.0","requestParameters":{"bucketName":"example-bucket","policy":"{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"s3:GetObject\",\"Resource\":\"arn:aws:s3:::example-bucket/*\"}]}"}},"responseElements":{"location":"http://example-bucket.s3.amazonaws.com/"},"requestID":"EXAMPLE123456789","eventID":"EXAMPLE-1-2-3-4-5-6","readOnly":false,"resources":[{"ARN":"arn:aws:s3:::example-bucket","accountId":"123456789012","type":"AWS::S3::Bucket"}],"eventType":"AwsApiCall","managementEvent":true,"recipientAccountId":"123456789012"}`
		msgs := []*message.Message{
			message.New().SetData([]byte(evt)),
			message.New().AsControl(),
		}

		fmt.Printf("input:\n%s\n", gjson.Get(evt, "@this|@pretty").String())
		fmt.Printf("output:\n")

		if _, err := sub.Transform(ctx, msgs...); err != nil {
			return err
		}

		fmt.Printf("\n")
		fmt.Printf("config:%s\n", demoConf)

		return nil
	},
}
