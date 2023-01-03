local sub = import '../../../../build/config/substation.libsonnet';

local processors = [
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  sub.patterns.processor.hash.data(set_key='event.created'),
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  sub.patterns.processor.time.now(set_key='event.hash'),
];

{
  processors: sub.helpers.flatten_processors(processors),
}
