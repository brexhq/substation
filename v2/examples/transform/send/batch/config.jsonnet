// This example configures send transforms with batch keys to organize
// data before it is sent externally. Every send transform supports batching
// and optionally grouping JSON objects by a value derived from the object.
local sub = import '../../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.object.copy({object: { source_key: '@pretty' }}),
    // Each object is organized by the value retrieved from the `group_id` key.
    sub.tf.send.stdout({object: { batch_key: 'group_id' }}),
    // This also applies to file-based send transforms, and every other send
    // transform as well.
    sub.tf.send.file({object: { batch_key: 'group_id' }}),
  ],
}
