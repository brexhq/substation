local process = import '../../../../build/config/process.libsonnet';

local processPatterns = import '../../../../build/config/process_patterns.libsonnet';

local event_created = 'event.created';
local event_hash = 'event.hash';

local processors = [
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  processPatterns.hash.data(set_key=event_hash),
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  process.apply(process.time(format='now'), set_key=event_created),
];

{
  processors: process.helpers.flatten_processors(processors),
}
