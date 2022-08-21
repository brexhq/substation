local conditionlib = import '../../../../build/config/condition.libsonnet';
local processlib = import '../../../../build/config/process.libsonnet';

local processors = [
  {
    // each row written to DynamoDB should be contained in an array stored in a single JSON key (e.g., __tmp.ddb); if the data is not an array, then Substation treats the key as an array of one item
    // the nested JSON key maps to the attribute name in DynamoDB:
    //  __tmp.ddb.PK maps to PK
    //  __tmp.ddb.SK maps to SK
    processors: [
      // copy the PK (required by the table)
      processlib.copy('event.hash', '__tmp.ddb.PK'),
      // insert the extra attributes
      processlib.copy('event.created', '__tmp.ddb.event_created'),
    ],
  },
  // if __tmp.ddb is empty, then drop the event to prevent the DynamoDB sink from processing unnecessary data
  {
    local conditions = [
      conditionlib.strings.empty('__tmp.ddb', negate=false),
    ],
    processors: [
      processlib.drop(condition_operator='and', condition_inspectors=conditions),
    ],
  },
];

{
  processors: std.flattenArrays([p.processors for p in processors]),
}
