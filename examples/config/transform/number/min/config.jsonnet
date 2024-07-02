// This example uses the `number_minimum` transform to return the smaller
// of two values, where one value is a constant and the other is a message.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.num.min({ value: 0 }),
    sub.tf.send.stdout(),
  ],
}
