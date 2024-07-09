// This example shows how to use the `utility_metric_bytes` transform to
// sum the amount of data received and transformed by Substation.
local sub = import '../../../../../build/config/substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
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
