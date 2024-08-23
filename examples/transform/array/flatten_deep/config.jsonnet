// This example flattens an array of arrays.
local sub = import '../../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // Flatten by copying the value and chaining GJSON's `@flatten` operator
    // with the `deep` option.
    sub.tf.object.copy({ object: { source_key: 'a|@flatten:{"deep":true}', target_key: 'a' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
