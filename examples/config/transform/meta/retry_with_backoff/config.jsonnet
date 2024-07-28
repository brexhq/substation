// This example shows how to implement retry with backoff behavior for any
// transform that does not produce an output. This technique may be useful
// when enriching data with external services or asynchronous data pipelines.
local sub = import '../../../../../build/config/substation.libsonnet';

// `key` is the target of the transform that may not produce an output and is
// checked to determine if the transform was successful.
local key = 'c';

local cnd = sub.cnd.all([
  sub.cnd.num.len.gt({ object: { source_key: key }, value: 0 }),
  sub.cnd.utility.random(),  // Simulates a transform that may fail to produce an output.
]);

{
  transforms: [
    sub.tf.meta.retry({
      transforms: [
        sub.tf.obj.insert({ object: { target_key: key }, value: true }),
      ],
      condition: cnd,  // If this returns false, then the transforms are retried.
      retry: { delay: '1s', count: 4 },  // Retry up to 4 times with a 1 second backoff (1s, 2s, 3s, 4s).
    }),
    sub.tf.send.stdout(),
  ],
}
