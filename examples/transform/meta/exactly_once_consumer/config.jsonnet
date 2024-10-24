// This example shows how to use the `meta_kv_store_lock` transform to
// create an "exactly once" semantic for a pipeline consumer.
local sub = import '../../../../substation.libsonnet';

// In production environments a distributed KV store should be used.
local kv = sub.kv_store.memory();

{
  tests: [
    {
      name: 'exactly_once_consumer',
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
    // If a message acquires a lock, then it is tagged for inspection.
    sub.tf.meta.kv_store.lock(settings={
      kv_store: kv,
      prefix: 'eo_consumer',
      ttl_offset: '1m',
      transforms: [
        sub.tf.obj.insert({ object: { target_key: 'meta eo_consumer' }, value: 'locked' }),
      ],
    }),
    // Messages that are not locked are dropped from the pipeline.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.none([
          sub.cnd.str.eq({ object: { source_key: 'meta eo_consumer' }, value: 'locked' }),
        ]),
        transforms: [
          sub.tf.utility.drop(),
        ],
      },
    ] }),
    // At this point only locked messages exist in the pipeline.
    sub.tf.send.stdout(),
  ],
}
