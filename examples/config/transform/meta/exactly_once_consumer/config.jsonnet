// This example shows how to use the `meta_kv_store_lock` transform to
// create an "exactly once" semantic for a pipeline consumer.
local sub = import '../../../../../build/config/substation.libsonnet';

// In production environments a distributed KV store should be used.
local kv = sub.kv_store.memory();

{
  transforms: [
    // If a message acquires a lock, then it is tagged for inspection.
    sub.tf.meta.kv_store.lock(settings={
      kv_store: kv,
      prefix: 'eo_consumer',
      ttl_offset: '1m',
      transform: sub.tf.obj.insert({ object: { target_key: 'meta eo_consumer' }, value: 'locked' }),
    }),
    // Messages that are not locked are dropped from the pipeline.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.none([
          sub.cnd.str.eq({ object: { source_key: 'meta eo_consumer' }, value: 'locked' }),
        ]),
        transform: sub.tf.utility.drop(),
      },
    ] }),
    // At this point only locked messages exist in the pipeline.
    sub.tf.send.stdout(),
  ],
}
