// This example shows how to use the `meta_kv_store_lock` transform to
// create an "exactly once" semantic for a pipeline producer.
local sub = import '../../../../../build/config/substation.libsonnet';

// In production environments a distributed KV store should be used.
local kv = sub.kv_store.memory();

{
  transforms: [
    // This only prints messages that acquire a lock. Any message
    // that fails to acquire a lock will be skipped. An error in the
    // sub-transform will cause all previously locked messages to be
    // unlocked.
    sub.tf.meta.err({ transform: sub.tf.meta.kv_store.lock(settings={
      kv_store: kv,
      prefix: 'eo_producer',
      ttl_offset: '1m',
      transform: sub.tf.send.stdout(),
    }) }),
  ],
}
