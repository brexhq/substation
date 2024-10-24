// This example flattens an array of arrays.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'extend',
      transforms: [
        sub.tf.test.message({ value: { a: [1, 2, [3, 4]] } }),
        sub.tf.send.stdout(),
      ],
      // Asserts that 'a' contains 4 elements.
      condition: sub.cnd.num.len.eq({ obj: { src: 'a' }, value: 4 }),
    },
  ],
  transforms: [
    // Flatten by copying the value and chaining GJSON's `@flatten` operator.
    sub.tf.obj.cp({ object: { source_key: 'a|@flatten', target_key: 'a' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
