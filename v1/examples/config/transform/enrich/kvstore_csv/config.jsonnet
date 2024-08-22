// This example shows how to use the `enrich_kv_store_item_get` transform
// to lookup data in a KV store backed by a CSV file.
local sub = import '../../../../../build/config/substation.libsonnet';

// This CSV file must be local to the Substation app. Absolute paths are
// recommended. Files accessible over HTTPS and hosted in AWS S3 also work.
//
// The `column` parameter is required and specifies the column in the CSV file
// that will be used to lookup the key in the KV store.
local kv = sub.kv_store.csv_file({ file: 'kv.csv', column: 'product' });

{
  transforms: [
    // The CSV file KV store returns the entire row minus the key column.
    // For example, this returns {"price":"9.99","calories":"500"} for "churro".
    sub.tf.enrich.kv_store.item.get({
      object: { source_key: 'product', target_key: 'info' },
      kv_store: kv,
    }),
    sub.tf.send.stdout(),
  ],
}
