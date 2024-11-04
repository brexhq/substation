package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/spf13/cobra"

	"github.com/brexhq/substation/v2"
)

var rootCmd = &cobra.Command{
	Use:  "substation",
	Long: "'substation' is a tool for managing Substation configurations.",
}

// transformRe captures the transform ID from a Substation error message.
// Example: `transform 324f1035-10a51b9a: object_target_key: missing required option` -> `324f1035-10a51b9a`
var transformRe = regexp.MustCompile(`transform ([a-f0-9-]+):`)

const (
	// confStdout is the default configuration used by
	// read-like commands. It prints any results (<= 100MB)
	// to stdout.
	confStdout = `local sub = std.extVar('sub');

{
  transforms: [
    sub.tf.send.stdout({ batch: { size: 100000000000, count: 1 } }),
  ],
}`

	// confDemo is a demo configuration for AWS CloudTrail.
	confDemo = `// Every config must import the Substation library.
local sub = std.extVar('sub');

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
}`

	evtDemo = `{"eventVersion":"1.08","userIdentity":{"type":"IAMUser","principalId":"EXAMPLE123456789","arn":"arn:aws:iam::123456789012:user/Alice","accountId":"123456789012","accessKeyId":"ASIAEXAMPLE123","sessionContext":{"attributes":{"mfaAuthenticated":"false","creationDate":"2024-10-01T12:00:00Z"},"sessionIssuer":{"type":"AWS","principalId":"EXAMPLE123456","arn":"arn:aws:iam::123456789012:role/Admin","accountId":"123456789012","userName":"Admin"}}},"eventTime":"2024-10-01T12:30:45Z","eventSource":"s3.amazonaws.com","eventName":"PutBucketPolicy","awsRegion":"us-west-2","sourceIPAddress":"203.0.113.0","userAgent":"aws-sdk-python/1.0.0 Python/3.8.0 Linux/4.15.0","requestParameters":{"bucketName":"example-bucket","policy":"{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"s3:GetObject\",\"Resource\":\"arn:aws:s3:::example-bucket/*\"}]}"}},"responseElements":{"location":"http://example-bucket.s3.amazonaws.com/"},"requestID":"EXAMPLE123456789","eventID":"EXAMPLE-1-2-3-4-5-6","readOnly":false,"resources":[{"ARN":"arn:aws:s3:::example-bucket","accountId":"123456789012","type":"AWS::S3::Bucket"}],"eventType":"AwsApiCall","managementEvent":true,"recipientAccountId":"123456789012"}`
)

func init() {
	// Hides the 'completion' command.
	rootCmd.AddCommand(&cobra.Command{
		Use:    "completion",
		Short:  "generate the autocompletion script for the specified shell",
		Hidden: true,
	})

	// Hides the 'help' command.
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

// compileFile returns JSON from a Jsonnet file.
func compileFile(fi string, extVars map[string]string) (string, error) {
	f, err := os.Open(fi)
	if err != nil {
		return "", err
	}
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return compileStr(string(s), extVars)
}

// compileStr returns JSON from a Jsonnet string.
func compileStr(s string, extVars map[string]string) (string, error) {
	vm := jsonnet.MakeVM()
	vm.ExtCode("sub", substation.Library)

	for k, v := range extVars {
		vm.ExtVar(k, v)
	}

	res, err := vm.EvaluateAnonymousSnippet("snippet", s)
	if err != nil {
		return "", err
	}

	return res, nil
}

// pathVars returns the directory and file name of a file path.
func pathVars(p string) (string, string) {
	dir, fn := filepath.Split(p)
	ext := filepath.Ext(fn)
	fn = strings.TrimSuffix(fn, ext)

	return dir, fn
}

// transformErrStr returns a formatted string for transform errors.
//
// If the error is not a transform error, then the error message
// is returned as is.
func transformErrStr(err error, arg string, cfg customConfig) string {
	r := transformRe.FindStringSubmatch(err.Error())

	// Cannot determine which transform failed. This should almost
	// never happen, unless something has modified the configuration
	// after it was compiled by Jsonnet.
	if len(r) == 0 {
		// Substation uses the transform name as a static transform ID.
		//
		// Example: `vet.json: transform hash_sha256: object_target_key: missing required option``
		return fmt.Sprintf("%s: %v\n", arg, err)
	}

	tfID := r[1] // The transform ID (e.g., `324f1035-10a51b9a`).

	// Prioritize returning test errors.
	for _, test := range cfg.Tests {
		for idx, tf := range test.Transforms {
			if tf.Settings["id"] == tfID {
				// Example: `vet.json:3 transform 324f1035-10a51b9a: object_target_key: missing required option``
				return fmt.Sprintf("%s:%d %v\n", arg, idx+1, err) + fmt.Sprintf("        %s\n\n", tf) // The line number is 1-based.
			}
		}
	}

	for idx, tf := range cfg.Config.Transforms {
		if tf.Settings["id"] == tfID {
			return fmt.Sprintf("%s:%d %v\n", arg, idx+1, err) + fmt.Sprintf("        %s\n\n", tf)
		}
	}

	// This happens if the input is not a transform error.
	return fmt.Sprintf("%s: %v\n", arg, err)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
