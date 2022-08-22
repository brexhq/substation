local conditionlib = import '../../../../build/config/condition.libsonnet';
local processlib = import '../../../../build/config/process.libsonnet';

local event_created = 'event.created';
local event_hash = 'event.hash';

local processors = [
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  {
    processors: [
      processlib.hash('@this', event_hash),
    ],
  },
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  {
    processors: [
      processlib.time('', event_created, 'now'),
    ],
  },
];

// flattens the `processors` array into a single array; required for compiling into config.jsonnet
{
  processors: std.flattenArrays([p.processors for p in processors]),
}
