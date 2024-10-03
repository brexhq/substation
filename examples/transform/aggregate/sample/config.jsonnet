// This example samples data by aggregating events into an array, then
// selecting the first event in the array as a sample. The sampling rate
// is 1/N, where N is the count of events in the buffer.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'sample',
      transforms: [
        sub.tf.test.message({ value: {"a":"b"}}),
        sub.tf.test.message({ value: {"c":"d"}}),
        sub.tf.test.message({ value: {"e":"f"}}),
        sub.tf.test.message({ value: {"g":"h"}}),
        sub.tf.test.message({ value: {"i":"j"}}),
        sub.tf.test.message({ value: {"k":"l"}}),
        sub.tf.test.message({ value: {"m":"n"}}),
        sub.tf.test.message({ value: {"o":"p"}}),
        sub.tf.test.message({ value: {"q":"r"}}),
        sub.tf.test.message({ value: {"s":"t"}}),
        sub.tf.test.message({ value: {"u":"v"}}),
        sub.tf.test.message({ value: {"w":"x"}}),
        sub.tf.test.message({ value: {"y":"z"}}),
        sub.tf.test.message({ value: " "}),
        sub.tf.send.stdout(),
      ],
      // Asserts that the message is '{"c":"d"}'.
      condition: sub.cnd.num.len.greater_than({ value: 0 }),
    }
  ],
  transforms: [
    // Events are aggregated into an array. This example has a sample
    // rate of up to 1/5. By default, the sample rate will be lower if
    // fewer than 5 events are processed by Substation.
    sub.tf.aggregate.to.array({ object: { target_key: 'meta sample' }, batch: { count: 5 } }),
    // A strict sample rate can be enforced by dropping any events that
    // contain the `sample` key, but do not have a length of 5.
    sub.tf.meta.switch(settings={ cases: [
      {
        condition: sub.cnd.num.len.eq({ object: { source_key: 'meta sample' }, value: 5 }),
        transforms: [
          sub.tf.object.copy({ object: { source_key: 'meta sample.0' } }),
        ],
      },
      {
        condition: sub.cnd.num.len.gt({ object: { source_key: 'meta sample' }, value: 0 }),
        transforms: [
          sub.tf.util.drop(),
        ],
      },
    ] }),
    sub.tf.obj.cp({ object: { source_key: 'meta sample.0' } }),
    sub.tf.send.stdout(),
  ],
}
