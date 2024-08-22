local sub = import '../../../../../build/config/substation.libsonnet';

local city = sub.kv_store.mmdb({ file: 'path/to/GeoLite2-City.mmdb' });

local asn = sub.kv_store.mmdb({ file: 'path/to/GeoLite2-ASN.mmdb' });

{
  transforms: [
    sub.tf.enrich.kv_store.iget({ object: { source_key: 'ip', target_key: 'city' }, kv_store: city }),
    sub.tf.enrich.kv_store.iget({ object: { source_key: 'ip', target_key: 'asn' }, kv_store: asn }),
    sub.tf.send.stdout(),
  ],
}
