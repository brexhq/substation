// This example configures custom retryable errors for the Kinesis Data Stream
// destination transform. All AWS transforms support a custom retry strategy,
// which can be used to handle transient errors in a way that is specific to
// the AWS service being used or the specific use case.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.aws.kinesis_data_stream(
      settings={ stream_name: 'substation', retry: {
        // The maximum number of times to retry a request.
        //
        // The default is 3.
        count: 3,
        // A list of regular expressions that match error messages
        // and cause the request to be retried. If there is no match, then
        // the default AWS retry strategy is used.
        //
        // The default is an empty list (i.e. no custom retryable errors).
        error_messages: ['connection reset by peer'],
      } },
    ),
  ],
}
