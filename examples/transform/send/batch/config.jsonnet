// This example configures send transforms with batch keys to organize
// data before it is sent externally. Every send transform supports batching
// and optionally grouping JSON objects by a value derived from the object.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'aux_transforms',
      transforms: [
        sub.tf.test.message({ value: { a: 'b', group_id: 1 } }),
        sub.tf.test.message({ value: { c: 'd', group_id: 2 } }),
        sub.tf.test.message({ value: { e: 'f', group_id: 1 } }),
        sub.tf.test.message({ value: { g: 'h', group_id: 2 } }),
        sub.tf.test.message({ value: { i: 'j', group_id: 1 } }),
        sub.tf.test.message({ value: { k: 'l', group_id: 2 } }),
        sub.tf.test.message({ value: { m: 'n', group_id: 1 } }),
        sub.tf.test.message({ value: { o: 'p', group_id: 2 } }),
        sub.tf.test.message({ value: { q: 'r', group_id: 1 } }),
        sub.tf.test.message({ value: { s: 't', group_id: 2 } }),
        sub.tf.test.message({ value: { u: 'v', group_id: 1 } }),
        sub.tf.test.message({ value: { w: 'x', group_id: 2 } }),
        sub.tf.test.message({ value: { y: 'z', group_id: 1 } }),
      ],
      condition: sub.cnd.num.len.gt({ value: 0 }),
    },
  ],
  transforms: [
    sub.tf.object.copy({ object: { source_key: '@pretty' } }),
    // Each object is organized by the value retrieved from the `group_id` key.
    sub.tf.send.stdout({ object: { batch_key: 'group_id' } }),
    // This also applies to file-based send transforms, and every other send
    // transform as well.
    sub.tf.send.file({ object: { batch_key: 'group_id' } }),
  ],
}
