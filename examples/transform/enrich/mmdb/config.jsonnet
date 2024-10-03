local sub = import '../../../../substation.libsonnet';

local asn = sub.kv_store.mmdb({ file: 'https://gist.github.com/jshlbrd/59641ccc71ba2873fb204ac44d101640/raw/3ad0e8c09563c614c50de4671caef8c1983cbb4d/GeoLite2-ASN.mmdb' });

{
  tests: [
    {
      name: 'mmdb-cloudflare',
      transforms: [
        sub.tf.test.message({ value: {"ip":"1.1.1.1"} }),
      ],
      // Asserts that the message contains ASN info.
      condition: sub.cnd.str.eq({ obj: {src: 'asn.autonomous_system_organization'}, value: 'CLOUDFLARENET' }),
    },
    {
      name: 'mmdb-google',
      transforms: [
        sub.tf.test.message({ value: {"ip":"8.8.8.8"} }),
      ],
      // Asserts that the message contains ASN info.
      condition: sub.cnd.str.eq({ obj: {src: 'asn.autonomous_system_organization'}, value: 'GOOGLE' }),
    }
  ],
  transforms: [
    sub.tf.enrich.kv_store.item.get({ object: { source_key: 'ip', target_key: 'asn' }, kv_store: asn }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
