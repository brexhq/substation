local consts = import 'consts.libsonnet';

local inspector = import '../../../../build/config/inspector.libsonnet';
local process = import '../../../../build/config/process.libsonnet';

local inspectorPatterns = import '../../../../build/config/inspector_patterns.libsonnet';
local operatorPatterns = import '../../../../build/config/operator_patterns.libsonnet';

local no_condition = {};

local processors = [
  {
    // each record written to DynamoDB should be put into in an array.
    // if the data is not an array, then the DynamoDB sink treats the value
    // as an array of one item.
    processors: [
      // copy the partition key (PK)
      process.apply(
        process.copy, key='event.hash', set_key=consts.ddb_payload + '.PK'),
      // insert the extra attributes
      process.apply(
        process.copy, key='event.created', set_key=consts.ddb_payload + '.event_created'),
    ],
  },
  // if !metadata ddb is empty, then drop the event to prevent the DynamoDB sink from processing unnecessary data
  {
    local gt_zero = inspector.inspect(
      options=inspectorPatterns.length.gt_zero, key=consts.ddb_payload,
    ),
    local op = operatorPatterns.and([gt_zero]),

    processors: [
      process.apply(options=process.drop, condition=op),
    ],
  },
];

{
  processors: std.flattenArrays([p.processors for p in processors]),
}
