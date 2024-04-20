local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  concurrency: 1,
  transforms: [
    // Puts the user's metadata into the KV store indexed by the host name.
    sub.tf.enrich.kv_store.set({ obj: { src: 'host.name', trg: 'user' }, prefix: 'md_user', kv_store: const.kv_store }),
  ],
}
