local inspector = import '../../build/config/inspector.libsonnet';
local process = import '../../build/config/process.libsonnet';

local inspectorPatterns = import '../../build/config/inspector_patterns.libsonnet';
local operatorPatterns = import '../../build/config/operator_patterns.libsonnet';
local processPatterns = import '../../build/config/process_patterns.libsonnet';

// inspectors and operators are combined to create unique matching patterns
//
// evalutes to true if the length of the value in "foo" is greater than zero
local foo_op = operatorPatterns.all([
  inspectorPatterns.length.gt_zero(key='foo'),
]);

// keys can be referenced outside of processor definitions
local event_created = 'event.created';
local event_hash = 'event.hash';

local processors = [
  // process patterns are put directly into the array.
  //
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  processPatterns.hash.data(set_key=event_hash),
  // processors are put directly into the array.
  //
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  process.apply(process.time(format='now'), set_key=event_created),
  // processors with local variables are objects in the array.
  //
  // if "foo" is not empty, then copy the value to "fu".
  {
    local output = 'fu',
    processor: process.apply(options=process.copy, key='foo', set_key=output, condition=foo_op),
  },
];

// nested processors are dynamically merged into a single array
{
  processors: process.helpers.flatten_processors(processors),
}
