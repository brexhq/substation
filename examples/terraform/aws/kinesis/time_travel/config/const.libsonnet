local sub = import '../../../../../../build/config/substation.libsonnet';

{
  is_process: [
    sub.cnd.str.eq({ obj: { src: 'event.category' }, value: 'process' }),
    sub.cnd.str.eq({ obj: { src: 'event.type' }, value: 'start' }),
  ],
  kv_store: sub.kv_store.aws_dynamodb({
    table_name: 'substation',
    attributes: { partition_key: 'PK', sort_key: 'SK', ttl: 'TTL', value: 'cache' },
  }),
}
