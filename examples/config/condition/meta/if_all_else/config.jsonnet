// This example determines if all values in an array are email addresses
// that have the DNS domain "brex.com". This technique can be used to
// validate or summarize values in an array.
local sub = import '../../../../../build/config/substation.libsonnet';

local any_gt_zero = sub.cnd.all(
  // After running the example, try changing this to "any" or "none" and see
  // what happens.
  sub.cnd.meta.for_each(settings={ type: 'none', inspector: sub.cnd.str.ends_with(settings={ string: '@brex.com' }) }),
);

{
  concurrency: 1,
  transforms: [
    sub.tf.send.stdout(),
    // In real-world scenarios, the match decision is typically used to
    // summarize an array of values. For this example, the decision is
    // represented as a boolean value and sent to stdout.
    sub.tf.meta.switch(
      settings={ switch: [
        {
          condition: any_gt_zero,
          transform: sub.tf.obj.insert(settings={ object: { set_key: 'meta result' }, value: true }),
        },
        {
          transform: sub.tf.obj.insert(settings={ object: { set_key: 'meta result' }, value: false }),
        },
      ] }
    ),
    sub.tf.obj.cp(settings={ object: { key: 'meta result' } }),
    sub.tf.send.stdout(),
  ],
}
