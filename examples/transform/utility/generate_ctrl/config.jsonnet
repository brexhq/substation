// This example shows how to use the `utility_control` transform to
// generate a control (ctrl) Message based on the amount of data Messages
// received by the system. ctrl Messages overrides the settings of the
// `aggregate_to_array` transform (and any other transform that supports).
local sub = std.extVar('sub');

{
  tests: [
    {
      name: 'generate_ctrl',
      transforms: [
        sub.tf.test.message({ value: { a: 'b' } }),
        sub.tf.test.message({ value: { c: 'd' } }),
        sub.tf.test.message({ value: { e: 'f' } }),
        sub.tf.test.message({ value: { g: 'h' } }),
        sub.tf.test.message({ value: { i: 'j' } }),
        sub.tf.test.message({ value: { k: 'l' } }),
        sub.tf.test.message({ value: { m: 'n' } }),
        sub.tf.test.message({ value: { o: 'p' } }),
        sub.tf.test.message({ value: { q: 'r' } }),
        sub.tf.test.message({ value: { s: 't' } }),
        sub.tf.test.message({ value: { u: 'v' } }),
        sub.tf.test.message({ value: { w: 'x' } }),
        sub.tf.test.message({ value: { y: 'z' } }),
        sub.tf.send.stdout(),
      ],
      // Asserts that the number of objects in each array is less than 3.
      condition: sub.cnd.num.len.less_than({ obj: { src: '@this' }, value: 3 }),
    },
  ],
  transforms: [
    sub.tf.utility.control({ batch: { count: 2 } }),
    sub.tf.aggregate.to.array(),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
