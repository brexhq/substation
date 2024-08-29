# Migration

Use this as a guide for migrating between major versions of Substation.

## v2.0.0

### Cmd

#### AWS Lambda Triggers

- Renamed `AWS_KINESIS_DATA_FIREHOSE` to `AWS_DATA_FIREHOSE`.
- Removed `AWS_KINESIS` (replaced by `AWS_KINESIS_DATA_STREAM`).
- Removed `AWS_DYNAMODB` (replaced by `AWS_DYNAMODB_STREAM`).

### Conditions

#### `meta.condition` Inspector

This is replaced by the `meta.all`, `meta.any`, and `meta.none` inspectors.

v1.x.x:

```jsonnet
sub.cnd.all([
  sub.cnd.str.eq({ value: 'FOO' }),
  sub.cnd.meta.condition({ condition: sub.cnd.any([
    sub.cnd.str.eq({ value: 'BAR' }),
    sub.cnd.str.eq({ value: 'BAZ' }),
  ]) }),
]),
```

v2.x.x:

```jsonnet
sub.cnd.all([
  sub.cnd.str.eq({ value: 'FOO' }),
  sub.cnd.any([
    sub.cnd.str.eq({ value: 'BAR' }),
    sub.cnd.str.eq({ value: 'BAZ' }),
  ]),
]),
```

#### `meta.for_each` Inspector

This is replaced by the `meta.all`, `meta.any`, and `meta.none` inspectors. If the `object.source_key` value is an array, then the data is treated as a list of elements. 

v1.x.x:

```jsonnet
sub.cnd.meta.for_each({
  object: { source_key: 'field' },
  type: 'any',
  inspector: sub.cnd.str.eq({ value: 'FOO' }),
})
```

v2.x.x:

```jsonnet
sub.cnd.meta.any([{
  object: { source_key: 'field' },
  inspectors: [ sub.cnd.str.eq({ value: 'FOO' }) ],
}])
```

#### `meta.negate` Inspector

This is replaced by the `meta.none` inspector.

v1.x.x:

```jsonnet
sub.cnd.meta.negate({ inspector: sub.cnd.str.eq({ value: 'FOO' }) })
```

v2.x.x:

```jsonnet
sub.cnd.meta.none({inspectors: [ sub.cnd.str.eq({ value: 'FOO' }) ]})
```


```jsonnet
sub.cnd.none([ sub.cnd.str.eq({ value: 'FOO' }) ])
```

#### `meta.err` Inspector

This is removed and was not replaced. Remove any references to this inspector.

### Transforms

#### `send.aws.*` Transforms

The AWS resource fields were replaced by an `aws` object field that contains the sub-fields `arn` and `assume_role_arn`. The region for each AWS client is derived from either the resource ARN or assumed role ARN.

v1.x.x:

```jsonnet
sub.tf.send.aws.s3({
  bucket_name: 'substation',
  file_path: { time_format: '2006/01/02/15', uuid: true, suffix: '.json' },
}),
```

v2.x.x:

```jsonnet
sub.tf.send.aws.s3({
  aws: { arn: 'arn:aws:s3:::substation' },
  file_path: { time_format: '2006/01/02/15', uuid: true, suffix: '.json' },
}),
```

**NOTE: This change also applies to every configuration that relies on an AWS resource.**

#### `meta.*` Transforms

The `transform` field is removed from all transforms and was replaced with the `transforms` field.

v1.x.x:

```jsonnet
sub.tf.meta.switch({ cases: [
  {
    condition: sub.cnd.all([
      sub.cnd.str.eq({ obj: { source_key: 'field' }, value: 'FOO' }),
    ]),
    transform: sub.tf.obj.insert({ object: { target_key: 'field' }, value: 'BAR' }),
  },
]})
```

v2.x.x:

```jsonnet
sub.tf.meta.switch({ cases: [
  {
    condition: sub.cnd.str.eq({ obj: { source_key: 'field' }, value: 'FOO' }),
    transforms: [ 
      sub.tf.obj.insert({ object: { target_key: 'field' }, value: 'BAR' }) 
    ],
  },
]})
```

#### `meta.retry` Transform

Retry settings were removed from all transforms and replaced by the `meta.retry` transform. It is recommended to create a reusable pattern for common retry scenarios.

v1.x.x:

```jsonnet
sub.tf.send.aws.sqs({
  aws: { arn: 'arn:aws:sqs:us-east-1:123456789012:my-queue' },
  retry: { count: 3 },
})
```

v2.x.x:

```jsonnet
sub.tf.meta.retry({
  retry: { count: 3, delay: '1s' },
  transforms: [
    sub.tf.send.aws.sqs({
      arn: 'arn:aws:sqs:us-east-1:123456789012:my-queue',
    }),
  ],
})
```

#### `meta.pipeline` Transform

This is removed and was not replaced. Remove any references to this transform and replace it with the `transforms` field used in other meta transforms.

#### `send.aws.dynamodb` Transform

The `send.aws.dynamodb` transform was renamed to `send.aws.dynamodb.put`.

v1.x.x:

```jsonnet
sub.tf.send.aws.dynamodb({
  table_name: 'substation',
}),
```

v2.x.x:

```jsonnet
sub.tf.send.aws.dynamodb.put({
  aws: { arn: 'arn:aws:dynamodb:us-east-1:123456789012:table/substation' },
}),
```

#### `enrich.aws.dynamodb` Transform

The `enrich.aws.dynamodb` transform was renamed to `enrich.aws.dynamodb.query`, and had these additional changes:
- `PartitionKey` and `SortKey` now reference the column names in the DynamoDB table and are nested under the `Attributes` field.
- By default, the value retrieved from `Object.SourceKey` is used as the `PartitionKey` value. If the `SortKey` is provided and the value from `Object.SourceKey` is an array, then the first element is used as the `PartitionKey` value and the second element is used as the `SortKey` value. 
- The `KeyConditionExpression` field was removed because this is now a derived value.

v1.x.x:

```jsonnet
// In v1.x.x, the DynamoDB column names must always be 'PK' and/or 'SK'.
sub.tf.obj.cp({ object: { src: key, trg: 'meta ddb.PK' } }),
sub.transform.enrich.aws.dynamodb({
  object: { source_key: 'meta ddb', target_key: 'user' },
  table_name: 'users_table',
  partition_key: 'PK',
  key_condition_expression: 'PK = :PK',
}),
```

v2.x.x:

```jsonnet
sub.transform.enrich.aws.dynamodb.query({
  object: { source_key: 'id', target_key: 'user' },
  aws: { arn: 'arn:aws:dynamodb:us-east-1:123456789012:table/users_table' },
  attributes: {
    partition_key: 'PK',
  },
}),
```

#### `send.aws.kinesis_data_firehose` Transform

The `send.aws.kinesis_data_firehose` transform was renamed to `send.aws.data_firehose`.

v1.x.x:

```jsonnet
sub.tf.send.aws.kinesis_data_firehose({
  stream_name: 'substation',
}),
```

v2.x.x:

```jsonnet
sub.tf.send.aws.data_firehose({
  aws: { arn: 'arn:aws:kinesis:us-east-1:123456789012:stream/substation' },
}),
```
