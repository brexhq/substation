// This example extends an array by appending and flattening values.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // Append the value of `z` to `a` (using the `-1` array index).
    sub.tf.object.copy({ object: { source_key: 'z', target_key: 'a.-1' } }),
    // Flatten the array.
    sub.tf.object.copy({ object: { source_key: 'a|@flatten', target_key: 'a' } }),
    sub.tf.send.stdout(),
  ],
}
