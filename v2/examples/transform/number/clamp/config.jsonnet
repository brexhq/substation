// This example shows how to clamp a number to a range.
local sub = import '../../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.number.maximum({ value: 0 }),
    sub.tf.number.minimum({ value: 100 }),
    sub.tf.send.stdout(),
  ],
}
