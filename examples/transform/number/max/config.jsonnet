// This example uses the `number_maximum` transform to return the larger
// of two values, where one value is a constant and the other is a message.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'max',
      transforms: [
        sub.tf.test.message({ value: 0 }),
        sub.tf.test.message({ value: 10 }),
        sub.tf.test.message({ value: -1 }),
        sub.tf.test.message({ value: -1.1 }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each number is greater than -1.
      condition: sub.cnd.num.greater_than({ value: -1 }),
    },
  ],
  transforms: [
    sub.tf.num.max({ value: 0 }),
    sub.tf.send.stdout(),
  ],
}
