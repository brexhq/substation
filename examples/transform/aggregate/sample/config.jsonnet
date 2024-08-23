// This example samples data by aggregating events into an array, then
// selecting the first event in the array as a sample. The sampling rate
// is 1/N, where N is the count of events in the buffer.
local sub = import '../../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // Events are aggregated into an array. This example has a sample
    // rate of up to 1/10. By default, the sample rate will be lower if
    // fewer than 10 events are processed by Substation.
    sub.tf.aggregate.to.array({ object: { target_key: 'meta sample' }, batch: { count: 10 } }),
    // A strict sample rate can be enforced by dropping any events that
    // contain the `sample` key, but do not have a length of 10.
    sub.tf.meta.switch(settings={ cases: [
      {
        condition: sub.cnd.num.len.eq({ object: { source_key: 'meta sample' }, value: 10 }),
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
