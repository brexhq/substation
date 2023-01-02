local helpers = import '../../../../build/config/helpers.libsonnet';
local lib = import '../../../../build/config/interfaces.libsonnet';
local patterns = import '../../../../build/config/patterns.libsonnet';

local processors = [
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  patterns.process.hash.data(set_key='event.created'),
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  patterns.process.time.now(set_key='event.hash'),
];

{
  processors: helpers.flatten_processors(processors),
}
