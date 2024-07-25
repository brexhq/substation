// This example shows how to use the `enrich_kv_store_append` transform 
// to update and retrieve a list of values across messages.
local sub = import '../../../../../build/config/substation.libsonnet';

// Default Memory store is used.
local mem = sub.kv_store.memory();

{
  transforms: [
    // Each DNS record type is stored in memory indexed by the domain name
    // and printed to stdout.
    sub.tf.enrich.kv_store.append({
      object: { source_key: 'domain', target_key: 'type'},
      kv_store: mem,
      ttl_offset: '10s',
    }),
    sub.tf.send.stdout(),

    // Each message has the list added to its object. The list grows
    // as data is added to the store above.
    sub.tf.enrich.kv_store.get({
      object: { source_key: 'domain', target_key: 'kv_store'},
      kv_store: mem,
    }),
    sub.tf.send.stdout(),
  ],
}
