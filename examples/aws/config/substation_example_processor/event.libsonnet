local process = import '../../../../build/config/process.libsonnet';

local processPatterns = import '../../../../build/config/process_patterns.libsonnet';

local event_created = 'event.created';
local event_hash = 'event.hash';

local processors =
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  processPatterns.hash.data(set_key=event_hash) + [
    // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
    {
      processors: [
        process.apply(
          process.time(format='now'), key='@this', set_key=event_created
        ),
      ],
    },
];

// flattens the `processors` array into a single array; required for compiling into config.jsonnet
{
  processors: std.flattenArrays([p.processors for p in processors]),
}
