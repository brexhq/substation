// This example uses the `number_minimum` transform to return the smaller
// of two values, where one value is a constant and the other is a message.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'min',
      transforms: [
        sub.tf.test.message({ value: 0 }),
        sub.tf.test.message({ value: 10 }),
        sub.tf.test.message({ value: -1 }),
        sub.tf.test.message({ value: -1.1 }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each number is less than 1.
      condition: sub.cnd.num.less_than({ value: 1 }),
    }
  ],
  transforms: [
    sub.tf.num.min({ value: 0 }),
    sub.tf.send.stdout(),
  ],
}
