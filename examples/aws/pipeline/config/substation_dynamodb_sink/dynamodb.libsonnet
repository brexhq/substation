local sub = import '../../../../../build/config/substation.libsonnet';

local const = import 'const.libsonnet';

// each record written to DynamoDB should be put into in an array.
// if the data is not an array, then the DynamoDB sink treats the value
// as an array of one item.
local processors = [
  // copy the partition key (PK)
  sub.interfaces.processor.copy(
    settings={key:'event.hash', set_key:sub.helpers.key.append(const.ddb_payload, 'PK')}
  ),
  // insert the extra attributes
  sub.interfaces.processor.copy(
    settings={key:'event.created', set_key:sub.helpers.key.append(const.ddb_payload, 'event_created')}
  ),
  // if !metadata ddb is empty, then drop the event to prevent the DynamoDB sink from processing unnecessary data
  sub.patterns.processor.if_not_empty(
    processor=sub.interfaces.processor.drop(), key=const.ddb_payload, negate=true
  ),
];

{
  processors: sub.helpers.flatten_processors(processors),
}
