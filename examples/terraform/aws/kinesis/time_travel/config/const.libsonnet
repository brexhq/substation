local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  kv_store: sub.kv_store.aws_dynamodb({
    table_name: 'substation',
    attributes: { partition_key: 'PK', sort_key: 'SK', ttl: 'TTL', value: 'cache' },
  }),
  field: 'context',
  field_exists: sub.cnd.num.len.gt({ obj: { src: $.field }, value: 0 }),
}
