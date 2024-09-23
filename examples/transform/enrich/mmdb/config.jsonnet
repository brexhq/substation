local sub = import '../../../../substation.libsonnet';

local city = sub.kv_store.mmdb({ file: 'https://gist.github.com/jshlbrd/59641ccc71ba2873fb204ac44d101640/raw/3ad0e8c09563c614c50de4671caef8c1983cbb4d/GeoLite2-City.mmdb' });

local asn = sub.kv_store.mmdb({ file: 'https://gist.github.com/jshlbrd/59641ccc71ba2873fb204ac44d101640/raw/3ad0e8c09563c614c50de4671caef8c1983cbb4d/GeoLite2-ASN.mmdb' });

{
  transforms: [
    sub.tf.enrich.kv_store.item.get({ object: { source_key: 'ip', target_key: 'city' }, kv_store: city }),
    sub.tf.enrich.kv_store.item.get({ object: { source_key: 'ip', target_key: 'asn' }, kv_store: asn }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
