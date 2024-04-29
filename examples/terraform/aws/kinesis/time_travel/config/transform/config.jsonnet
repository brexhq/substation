// All values in the KV store were put there by the `enrichment` function.
local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  concurrency: 2,
  transforms: [
    // process.*
    //
    // This is only applied to non-process events.
    sub.pattern.tf.conditional(
      condition=sub.cnd.none(const.is_process),
      transform=sub.tf.enrich.kv_store.get({ obj: { src: 'process.pid', trg: 'process' }, prefix: 'process', kv_store: const.kv_store }),
    ),
    // process.parent.*
    sub.pattern.tf.conditional(
      condition=sub.cnd.num.len.gt({ obj: { src: 'process.parent.pid' }, value: 0 }),
      transform=sub.tf.enrich.kv_store.get({ obj: { src: 'process.parent.pid', trg: 'process.parent' }, prefix: 'process', kv_store: const.kv_store }),
    ),
    // process.parent.parent.*
    sub.pattern.tf.conditional(
      condition=sub.cnd.num.len.gt({ obj: { src: 'process.parent.parent.pid' }, value: 0 }),
      transform=sub.tf.enrich.kv_store.get({ obj: { src: 'process.parent.parent.pid', trg: 'process.parent.parent' }, prefix: 'process', kv_store: const.kv_store }),
    ),
    // Print the results.
    sub.tf.send.stdout(),
  ],
}
