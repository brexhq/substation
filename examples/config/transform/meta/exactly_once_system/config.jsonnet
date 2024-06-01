// This example shows how to use the `meta_kv_store_lock` transform to
// create an "exactly once" semantic for an entire pipeline system.
local sub = import '../../../../../build/config/substation.libsonnet';

// In production environments a distributed KV store should be used.
local kv = sub.kv_store.memory();

{
  transforms: [
    // All messages are locked before being sent through other transform
    // functions, ensuring that the message is processed only once.
    // An error in any sub-transform will cause all previously locked
    // messages to be unlocked.
    sub.tf.meta.err({ transform: sub.tf.meta.kv_store.lock(settings={
      kv_store: kv,
      prefix: 'eo_system',
      ttl_offset: '1m',
      transform: sub.tf.meta.pipeline({ transforms: [
        sub.tf.obj.insert({ object: { target_key: 'processed' }, value: true }),
        sub.tf.send.stdout(),
      ] }),
    }) }),
  ],
}
