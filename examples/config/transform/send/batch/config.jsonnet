// This example configures send transforms with batch keys to organize
// data before it is sent externally. Every send transform supports batching
// and optionally grouping JSON objects by a value derived from the object.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.stdout({
      // Each object is organized by the value retrieved from the `group_id` key.
      object: { batch_key: 'group_id' },
    }),
    sub.tf.send.file({
      // This also applies to file-based send transforms, and every other send
      // transform as well.
      object: { batch_key: 'group_id' },
    }),
  ],
}
