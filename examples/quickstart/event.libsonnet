local sub = import '../../build/config/substation.libsonnet';
// nested defintions can be accessed after importing sub ...
// local help = sub.helpers;
// local insp = sub.interfaces.inspector;
// local proc = sub.interfaces.processor;

// inspectors and operators are combined to create unique matching
// patterns. interfaces and patterns can be combined as needed.
//
// evalutes to true if the length of the value in "foo" is greater
// than zero. this operator is automated in 
// sub.patterns.process.if_not_empty pattern.
local foo_op = sub.interfaces.operator.all(
  sub.patterns.inspector.length.gt_zero(key='foo'),
);

// keys can be referenced outside of processor definitions
local event_created = 'event.created';
local event_hash = 'event.hash';

local processors = [
  // process patterns are put directly into the array.
  //
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  sub.patterns.processor.hash.data(set_key=event_hash),
  // processors are put directly into the array. this is identical to 
  // sub.patterns.processor.time.now().
  //
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
  sub.interfaces.processor.time(options={format:'now'}, settings={set_key:event_created}),
  // processors with local variables are objects in the array.
  //
  // if "foo" is not empty, then copy the value to "fu".
  {
    local s = {key:'foo', set_key:'fu', condition:foo_op},
    processor: sub.interfaces.processor.copy(settings=s)
  },
];

// processors are dynamically merged into a single array
{
  processors: sub.helpers.flatten_processors(processors),
}
