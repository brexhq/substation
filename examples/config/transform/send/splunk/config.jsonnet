// This example transforms data into a Splunk HTTP Event Collector (HEC)
// compatible format and sends it to a Splunk instance. The Splunk HEC
// expects mulitple events to be sent in a single request using this format:
// {"a":"b"}{"c":"d"}{"e":"f"}
//
// More information about the Splunk HEC can be found here:
// https://docs.splunk.com/Documentation/SplunkCloud/latest/Data/HECExamples
local sub = import '../../../../../../build/config/substation.libsonnet';

// By default the Splunk HEC limits the size of each request to 1MB.
local max_size = 1000 * 1000;

{
  concurrency: 1,
  transforms: [
    // Connections to the Splunk HEC are authenticated using a token.
    sub.transform.utility.secret({ secret: sub.secrets.environment_variable({ id: 'SPLUNK', name: 'SPLUNK_TOKEN_ID' }) }),
    sub.tf.send.http.post({
      batch: { size: max_size },
      auxiliary_transforms: [
        sub.tf.agg.to.array(),
        sub.tf.array.join({ separator: '' }),
      ],
      url: 'https://my-instance.cloud.splunk.com:8088/services/collector',
      headers: [{
        key: 'Authorization',
        value: 'Splunk ${SECRET:SPLUNK}',
      }],
    }),
  ],
}
