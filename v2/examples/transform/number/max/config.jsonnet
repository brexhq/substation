// This example uses the `number_maximum` transform to return the larger
// of two values, where one value is a constant and the other is a message.
local sub = import '../../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.num.max({ value: 0 }),
    sub.tf.send.stdout(),
  ],
}
