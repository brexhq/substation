local helpers = import '../../../../build/config/helpers.libsonnet';
local lib = import '../../../../build/config/interfaces.libsonnet';
local patterns = import '../../../../build/config/patterns.libsonnet';

local consts = import 'consts.libsonnet';

// each record written to DynamoDB should be put into in an array.
// if the data is not an array, then the DynamoDB sink treats the value
// as an array of one item.
local processors = [
  // copy the partition key (PK)
  lib.process.apply(
    lib.process.copy,
    key='event.hash',
    set_key=helpers.key.append(consts.ddb_payload, 'PK'),
  ),
  // insert the extra attributes
  lib.process.apply(
    lib.process.copy,
    key='event.created',
    set_key=helpers.key.append(consts.ddb_payload, 'event_created'),
  ),
  // if !metadata ddb is empty, then drop the event to prevent the DynamoDB sink from processing unnecessary data
  patterns.process.if_not_empty(options=lib.process.drop, key=consts.ddb_payload, negate=true),
];

{
  processors: helpers.flatten_processors(processors),
}
