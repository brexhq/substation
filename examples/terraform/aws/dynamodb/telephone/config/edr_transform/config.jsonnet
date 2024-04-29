local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

// cnd_copy is a helper function for copying values that are not null.
local cnd_copy(source, target) = sub.pattern.tf.conditional(
  condition=sub.cnd.num.len.gt({ obj: { src: source }, value: 0 }),
  transform=sub.tf.object.copy({ obj: { src: source, trg: target } }),
);

{
  concurrency: 1,
  transforms: [
    // The value from the KV store can be null, so the result is hidden in metadata and checked before
    // copying it into the message data. Many of these values are supersets of each other, so values are
    // overwritten if they exist. If any source key is missing, the transform is skipped.
    sub.tf.enrich.kv_store.get({ obj: { src: 'host.id', trg: 'meta edr_host' }, prefix: 'edr_host', kv_store: const.kv_store }),
    cnd_copy(source='meta edr_host', target='host'),
    sub.tf.enrich.kv_store.get({ obj: { src: 'host.name', trg: 'meta md_user' }, prefix: 'md_user', kv_store: const.kv_store }),
    cnd_copy(source='meta md_user', target='user'),
    sub.tf.enrich.kv_store.get({ obj: { src: 'user.email', trg: 'meta idp_user' }, prefix: 'idp_user', kv_store: const.kv_store }),
    cnd_copy(source='meta idp_user', target='user'),
    sub.tf.send.stdout(),
  ],
}
