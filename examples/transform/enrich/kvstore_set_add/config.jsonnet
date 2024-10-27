// This example shows how to use the `enrich_kv_store_set_add` transform
// to track data over time in a KV store. The sample data contains food
// orders and is indexed by each customer's email address.
local sub = std.extVar('sub');

// Default Memory store is used.
local mem = sub.kv_store.memory();

{
  tests: [
    {
      name: 'kvstore_set_add',
      transforms: [
        sub.tf.test.message({ value: { date: '2021-01-01', customer: 'alice@brex.com', order: 'pizza' } }),
        sub.tf.test.message({ value: { date: '2021-01-01', customer: 'bob@brex.com', order: 'burger' } }),
        sub.tf.test.message({ value: { date: '2021-01-03', customer: 'bob@brex.com', order: 'pizza' } }),
        sub.tf.test.message({ value: { date: '2021-01-07', customer: 'alice@brex.com', order: 'pizza' } }),
        sub.tf.test.message({ value: { date: '2021-01-07', customer: 'bob@brex.com', order: 'burger' } }),
        sub.tf.test.message({ value: { date: '2021-01-13', customer: 'alice@brex.com', order: 'pizza' } }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each message is not empty.
      condition: sub.cnd.num.len.gt({ value: 0 }),
    },
  ],
  transforms: [
    // Each order is stored in memory indexed by the customer's email
    // address and printed to stdout. Only unique orders are stored.
    sub.tf.enrich.kv_store.set.add({
      object: { source_key: 'customer', target_key: 'order' },
      kv_store: mem,
      ttl_offset: '10s',
    }),
    // Each message has the list added to its object. The list grows
    // as orders are added to the store above.
    sub.tf.enrich.kv_store.item.get({
      object: { source_key: 'customer', target_key: 'kv_store' },
      kv_store: mem,
    }),
    sub.tf.send.stdout(),
  ],
}
