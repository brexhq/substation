// This example shows how to use the `meta_kv_store_lock` transform to
// execute a transform exactly once within a time window. The transform
// will print unique values from `./data.jsonl` once to stdout and on
// subsequent runs will not print anything until the TTL expires.
local sub = import '../../../../../../build/config/substation.libsonnet';

// DynamoDB is used as the store for maintaining lock state. The table 
// must have TTL enabled. Required attributes depend on the table schema.
local kv = sub.kv_store.aws_dynamodb(settings={
	table_name: 'substation',
	attributes: {
		partition_key: 'PK',
		sort_key: 'SK',
		ttl: 'TTL',
	},
});

{
  transforms: [
	// This transform will print unique values to stdout exactly once within
	// a time window of 1 minute.
	sub.tf.meta.kv_store.lock(settings={
		transform: sub.tf.send.stdout(),
		prefix: 'exec_once',
		ttl_offset: '1m',
		kv_store: kv,
	})
  ],
}
