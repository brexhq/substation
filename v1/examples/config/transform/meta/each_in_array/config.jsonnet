// This example shows how to use the `meta.for_each` transform to
// modify objects in an array. In this example, keys are removed
// and added to each object in the array.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.tf.meta.for_each({
      object: { source_key: 'a', target_key: 'a' },
      // Multiple transforms can be applied in series to each object
      // in the array by using the `meta.pipeline` transform. Otherwise,
      // use any individual transform to modify the object.
      transform: sub.tf.meta.pipeline({ transforms: [
        sub.tf.object.delete({ object: { source_key: 'b' } }),
        sub.tf.object.insert({ object: { target_key: 'z' }, value: true }),
      ] }),
    }),
    sub.tf.send.stdout(),
  ],
}
