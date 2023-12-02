// This example shows how to use the `utility_metric_rate` transform to
// measure the rate of messages processed by Substation.
local sub = import '../../../../../build/config/substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
  transforms: [
    // If the transform is configured first, then the count reflects
    // the number of messages received by Substation.
    sub.transform.utility.metric.rate(
      settings={ metric: { name: 'MessageThroughput', attributes: attr, destination: dest } },
    ),
  ],
}
