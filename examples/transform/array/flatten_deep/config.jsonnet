// This example flattens an array of arrays.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'extend',
      transforms: [
        sub.tf.test.message({ value: { a: [1, 2, [3, 4, [5, 6]]] } }),
        sub.tf.send.stdout(),
      ],
      // Asserts that 'a' contains 6 elements.
      condition: sub.cnd.num.len.eq({ obj: { src: 'a' }, value: 6 }),
    },
  ],
  transforms: [
    // Flatten by copying the value and chaining GJSON's `@flatten` operator
    // with the `deep` option.
    sub.tf.object.copy({ object: { source_key: 'a|@flatten:{"deep":true}', target_key: 'a' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
