// This example shows how to use the `utility_metric_bytes` transform to
// sum the amount of data received and transformed by Substation.
local sub = import '../../../../substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
  tests: [
    {
      name: 'message_bytes',
      transforms: [
        sub.tf.test.message({ value: {"a":"b"}}),
        sub.tf.test.message({ value: {"c":"d"}}),
        sub.tf.test.message({ value: {"e":"f"}}),
        sub.tf.test.message({ value: {"g":"h"}}),
        sub.tf.test.message({ value: {"i":"j"}}),
        sub.tf.test.message({ value: {"k":"l"}}),
        sub.tf.test.message({ value: {"m":"n"}}),
        sub.tf.test.message({ value: {"o":"p"}}),
        sub.tf.test.message({ value: {"q":"r"}}),
        sub.tf.test.message({ value: {"s":"t"}}),
        sub.tf.test.message({ value: {"u":"v"}}),
        sub.tf.test.message({ value: {"w":"x"}}),
        sub.tf.test.message({ value: {"y":"z"}}),
      ],
      // Asserts that each message is not empty.
      condition: sub.cnd.num.len.gt({ value: 0 }),
    }
  ],
  transforms: [
    // If the transform is configured first, then the metric reflects
    // the sum of bytes received by Substation.
    sub.transform.utility.metric.bytes({ metric: { name: 'BytesReceived', attributes: attr, destination: dest } }),
    // This inserts a value into the object so that the message size increases.
    sub.transform.object.insert({ obj: { target_key: '_' }, value: 1 }),
    // If the transform is configured last, then the metric reflects
    // the sum of bytes transformed by Substation.
    sub.transform.utility.metric.bytes({ metric: { name: 'BytesTransformed', attributes: attr, destination: dest } }),
  ],
}
