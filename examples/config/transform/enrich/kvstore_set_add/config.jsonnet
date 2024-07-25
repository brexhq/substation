// This example shows how to use the `enrich_kv_store_set_add` transform 
// to track data over time in a KV store. The sample data contains food
// orders and is indexed by each customer's email address.
local sub = import '../../../../../build/config/substation.libsonnet';

// Default Memory store is used.
local mem = sub.kv_store.memory();

{
  transforms: [
    // Each order is stored in memory indexed by the customer's email
    // address and printed to stdout. Only unique orders are stored.
    sub.tf.enrich.kv_store.set.add({
      object: { source_key: 'customer', target_key: 'order'},
      kv_store: mem,
      ttl_offset: '10s',
    }),
    sub.tf.send.stdout(),

    // Each message has the list added to its object. The list grows
    // as orders are added to the store above.
    sub.tf.enrich.kv_store.item.get({
      object: { source_key: 'customer', target_key: 'kv_store'},
      kv_store: mem,
    }),
    sub.tf.send.stdout(),
  ],
}
