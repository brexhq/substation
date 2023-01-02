local helpers = import '../../build/config/helpers.libsonnet';
local lib = import '../../build/config/interfaces.libsonnet';
local patterns = import '../../build/config/patterns.libsonnet';

// inspectors and operators are combined to create unique matching
// patt. interfaces and patterns can be combined if needed.
//
// evalutes to true if the length of the value in "foo" is greater
// than zero. this operator is automated in the process.if_not_empty
// pattern.
local foo_op = lib.operator.all(
  patterns.inspect.length.gt_zero(key='foo'),
);

// keys can be referenced outside of processor definitions
local event_created = 'event.created';
local event_hash = 'event.hash';

local processors = [
  // process patterns are put directly into the array.
  //
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  patterns.process.hash.data(set_key=event_hash),
  // processors are put directly into the array. this is identical to the
  // process.time.now pattern.
  //
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  lib.process.apply(lib.process.time(format='now'), set_key=event_created),
  // processors with local variables are objects in the array.
  //
  // if "foo" is not empty, then copy the value to "fu".
  {
    local opts = lib.process.copy,
    processor: lib.process.apply(options=opts, key='foo', set_key='fu', condition=foo_op),
  },
];

// nested processors are dynamically merged into a single array
{
  processors: helpers.flatten_processors(processors),
}
