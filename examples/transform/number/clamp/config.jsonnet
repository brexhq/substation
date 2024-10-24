// This example shows how to clamp a number to a range.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'clamp',
      transforms: [
        sub.tf.test.message({ value: -1 }),
        sub.tf.test.message({ value: 101 }),
        sub.tf.test.message({ value: 50 }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each number is within the range [0, 100].
      condition: sub.cnd.all([
        sub.cnd.num.greater_than({ value: -1 }),
        sub.cnd.num.less_than({ value: 101 }),
      ]),
    },
  ],
  transforms: [
    sub.tf.number.maximum({ value: 0 }),
    sub.tf.number.minimum({ value: 100 }),
    sub.tf.send.stdout(),
  ],
}
