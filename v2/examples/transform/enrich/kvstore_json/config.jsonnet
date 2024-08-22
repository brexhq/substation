// This example shows how to use the `enrich_kv_store_item_get` transform 
// to lookup data in a KV store backed by a JSON file.
local sub = import '../../../../substation.libsonnet';

// This JSON file must be local to the Substation app. Absolute paths are
// recommended. Files accessible over HTTPS and hosted in AWS S3 also work.
local kv = sub.kv_store.json_file({ file: 'kv.json' });

{
  transforms: [
    sub.tf.enrich.kv_store.item.get({
      object: { source_key: 'product', target_key: 'price'},
      kv_store: kv,
    }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
