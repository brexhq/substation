local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  concurrency: 2,
  transforms: [
    // If the field doesn't exist, then get the value from the KV store.
    // The value should have been previously placed into the store by the
    // enrichment node.
    sub.pattern.tf.conditional(
      condition=sub.cnd.none(const.field_exists),
      transform=sub.tf.enrich.kv_store.get({ obj: { src: 'ip', trg: const.field }, kv_store: const.kv_store }),
    ),
    sub.tf.send.stdout(),
  ],
}
