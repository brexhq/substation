// This config transforms the failure record sent by the node Lambda function
// so that it becomes a new request. The new request bypasses S3 and is sent
// directly to the Lambda function.
//
// Additional information is available in the payload and can be used to make
// decisions about the new request or notify external systems about the failure.
local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // If needed, then use other information from the failure record to
    // decide what to do or notify external systems about the failure.
    sub.tf.obj.cp(settings={ object: { source_key: 'requestPayload' } }),
    sub.tf.send.aws.lambda(settings={ function_name: 'node' }),
  ],
}
