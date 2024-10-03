// This example extends an array by appending and flattening values.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'extend',
      transforms: [
        sub.tf.test.message({ value: {"a":[1,2],"z":[3,4]} }),
        sub.tf.send.stdout(),
      ],
      // Asserts that 'a' contains 4 elements.
      condition: sub.cnd.num.len.eq({ obj: {src: 'a'}, value: 4 }),
    }
  ],
  transforms: [
    // Append the value of `z` to `a` (using the `-1` array index).
    sub.tf.object.copy({ object: { source_key: 'z', target_key: 'a.-1' } }),
    // Flatten the array.
    sub.tf.object.copy({ object: { source_key: 'a|@flatten', target_key: 'a' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
