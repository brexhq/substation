// This example shows how to use the `utility_metric_count` transform to
// count the number of messages received and transformed by Substation.
local sub = import '../../../../substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
  tests: [
    {
      name: 'message_count',
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
    // If the transform is configured first, then the count reflects
    // the number of messages received by Substation.
    sub.transform.utility.metric.count({ metric: { name: 'MessagesReceived', attributes: attr, destination: dest } }),
    sub.transform.utility.drop(),
    // If the transform is configured last, then the count reflects
    // the number of messages transformed by Substation.
    sub.transform.utility.metric.count({ metric: { name: 'MessagesTransformed', attributes: attr, destination: dest } }),
  ],
}
