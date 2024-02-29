local sub = import '../../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  // The concurrency is set to 1 to ensure that the KV store is not updated in parallel.
  concurrency: 1,
  transforms: [
    // If the field exists, then put the value into the KV store. If the data stream is
    // at risk of write heavy activity, then consider first querying the KV store to see
    // if the value already exists and only writing if it does not.
    sub.pattern.tf.conditional(
      condition=sub.cnd.all(const.field_exists),
      // The ttl_offset is low for the purposes of this example. It should be set to a
      // value that is appropriate for the data stream (usually hours or days).
      transform=sub.tf.enrich.kv_store.set({ obj: { src: 'ip', trg: const.field }, ttl_offset: '30s', kv_store: const.kv_store }),
    ),
  ],
}
