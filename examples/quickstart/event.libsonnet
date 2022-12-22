local condition = import '../../build/config/condition.libsonnet';
local process = import '../../build/config/process.libsonnet';

local inspectorPatterns = import '../../build/config/inspector_patterns.libsonnet';
local operatorPatterns = import '../../build/config/operator_patterns.libsonnet';
local processPatterns = import '../../build/config/process_patterns.libsonnet';

// inspectors and operators are combined to create unique matching patterns
//
// evalutes to true if the length of the value in "foo" is greater than zero
local foo_gt_zero = condition.inspector(inspectorPatterns.length.gt_zero, key='foo');
local foo_op = operatorPatterns.and([foo_gt_zero]);

// keys can be referenced outside of processor definitions
local event_created = 'event.created';
local event_hash = 'event.hash';

local processors =
  // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-hash
  processPatterns.hash.data(set_key=event_hash) +
  [
    // if "foo" is not empty, then copy the value to "fu"
    {
      processors: [
        process.process(process.copy, key='foo', set_key='fu', condition=foo_op),
      ],
    },
    // generates current time and automatically formats it
    // https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-created
    {
      processors: [
        process.process(process.time(format='now'), key='@this', set_key=event_created),
      ],
    },
  ];

{
  processors: std.flattenArrays([p.processors for p in processors]),
}
