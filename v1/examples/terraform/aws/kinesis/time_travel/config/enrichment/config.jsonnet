// Puts process metadata into the KV store.
local sub = import '../../../../../../../build/config/substation.libsonnet';
local const = import '../const.libsonnet';

{
  // The concurrency is set to 1 to ensure that the KV store is not updated in parallel.
  concurrency: 1,
  transforms: [
    // If the event is a process, then store the process metadata in the KV store
    // indexed by the PID. The data is stored in the KV store for 90 days.
    sub.pattern.tf.conditional(
      condition=sub.cnd.all(const.is_process),
      transform=sub.tf.enrich.kv_store.iset({ obj: { src: 'process.pid', trg: 'process' }, prefix: 'process', ttl_offset: std.format('%dh', 24 * 90), kv_store: const.kv_store, close_kv_store: false }),
    ),
  ],
}
