local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  concurrency: 1,
  transforms: [
    // If the host metadata contains the host name, then it's put into the KV store
    // indexed by the host ID.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.all([
          sub.cnd.num.len.gt({ obj: { src: 'host.name' }, value: 0 }),
        ]),
        transform: sub.tf.enrich.kv_store.iset({ obj: { src: 'host.id', trg: 'host' }, prefix: 'edr_host', kv_store: const.kv_store }),
      },
    ] }),
  ],
}
