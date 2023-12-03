local sub = import '../../../../../../build/config/substation.libsonnet';

local city = sub.kv_store.mmdb(
  settings={ file: 'path/to/GeoLite2-City.mmdb' }
);

local asn = sub.kv_store.mmdb(
  settings={ file: 'path/to/GeoLite2-ASN.mmdb' }
);

{
  transforms: [
    sub.transform.enrich.kv_store.get(
      settings={ object: { key: 'ip', set_key: 'city' }, kv_store: city }
    ),
    sub.tf.enrich.kv_store.get(
      settings={ object: { key: 'ip', set_key: 'asn' }, kv_store: asn }
    ),
    sub.tf.send.stdout(),
  ],
}
