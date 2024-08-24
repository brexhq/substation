// This example shows how to use the `utility_metric_freshness` transform to
// determine if data was processed by the system within a certain time frame.
//
// Freshness is calculated by comparing a time value in the message to the current
// time and determining if the difference is less than a threshold:
// - Success: current time - timestamp < threshold
// - Failure: current time - timestamp >= threshold
//
// The transform emits two metrics that describe success and failure, annotated
// in the `FreshnessType` attribute.
local sub = import '../../../../substation.libsonnet';

local attr = { AppName: 'example' };
local dest = { type: 'aws_cloudwatch_embedded_metrics' };

{
  transforms: [
    sub.transform.utility.metric.freshness({
      threshold: '5s',  // Amount of time spent in the system before considered stale.
      object: { source_key: 'timestamp' },  // Used as the reference to determine freshness.
      metric: { name: 'MessageFreshness', attributes: attr, destination: dest },
    }),
  ],
}
