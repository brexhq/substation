// This example shows how to use the `meta_kv_store_lock` transform to
// create an "exactly once" semantic for a pipeline producer.
local sub = import '../../../../substation.libsonnet';

// In production environments a distributed KV store should be used.
local kv = sub.kv_store.memory();

{
  tests: [
    {
      name: 'exactly_once_producer',
      transforms: [
        sub.tf.test.message({ value: { a: 'b' } }),
        sub.tf.test.message({ value: { a: 'b' } }),
        sub.tf.test.message({ value: { c: 'd' } }),
        sub.tf.test.message({ value: { a: 'b' } }),
        sub.tf.test.message({ value: { c: 'd' } }),
        sub.tf.test.message({ value: { c: 'd' } }),
        sub.tf.test.message({ value: { e: 'f' } }),
        sub.tf.test.message({ value: { a: 'b' } }),
      ],
      // Asserts that each message is not empty.
      condition: sub.cnd.num.len.gt({ value: 0 }),
    },
  ],
  transforms: [
    // This only prints messages that acquire a lock. Any message
    // that fails to acquire a lock will be skipped. An error in the
    // sub-transform will cause all previously locked messages to be
    // unlocked.
    sub.tf.meta.err({ transforms: [
      sub.tf.meta.kv_store.lock({
        kv_store: kv,
        prefix: 'eo_producer',
        ttl_offset: '1m',
        transforms: [
          sub.tf.send.stdout(),
        ],
      }),
    ] }),
  ],
}
