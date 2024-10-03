// This example shows how to use the `meta.for_each` transform to
// modify objects in an array. In this example, keys are removed
// and added to each object in the array.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'each_in_array',
      transforms: [
        sub.tf.test.message({ value: {"a":[{"b":1,"c":2},{"b":3,"c":4}]} }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each object in the array 'a' has:
      //  - No key 'b'
      //  - Key 'z' with value 'true'
      condition: sub.cnd.meta.all({
        object: { src: 'a' },
        conditions: [
          sub.cnd.num.len.eq({ obj: {src: 'b'}, value: 0 }),
          sub.cnd.str.eq({ obj: {src: 'z'}, value: 'true' }),
        ],
      }),
    }
  ],
  transforms: [
    sub.tf.meta.for_each({
      object: { source_key: 'a', target_key: 'a' },
      // Multiple transforms can be applied in series to each object
      // in the array by using the `meta.pipeline` transform. Otherwise,
      // use any individual transform to modify the object.
      transforms: [
        sub.tf.object.delete({ object: { source_key: 'b' } }),
        sub.tf.object.insert({ object: { target_key: 'z' }, value: true }),
      ],
    }),
    sub.tf.send.stdout(),
  ],
}
