# Migration

Use this as a guide for migrating between major versions of Substation.

## v2.0.0

### Applications (cmd/)

#### AWS Lambda Handlers

Multiple AWS Lambda handlers were renamed to better reflect the AWS service they interact with:
- Renamed `AWS_KINESIS_DATA_FIREHOSE` to `AWS_DATA_FIREHOSE`.
- Renamed `AWS_KINESIS` to `AWS_KINESIS_DATA_STREAM`.
- Renamed `AWS_DYNAMODB` to `AWS_DYNAMODB_STREAM`.

v1.x.x:

```hcl
module "node" {
  source    = "build/terraform/aws/lambda"

  config = {
    name        = "node"
    description = "Substation node that is invoked by a Kinesis Data Stream."
    image_uri   = "123456789012.dkr.ecr.us-east-1.amazonaws.com/substation:v1.0.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS"
    }
  }
}
```

v2.x.x:

```hcl
module "node" {
  source    = "build/terraform/aws/lambda"

  config = {
    name        = "node"
    description = "Substation node that is invoked by a Kinesis Data Stream."
    image_uri   = "123456789012.dkr.ecr.us-east-1.amazonaws.com/substation:v2.0.0"
    image_arm   = true

    env = {
      "SUBSTATION_CONFIG" : "http://localhost:2772/applications/substation/environments/example/configurations/node"
      "SUBSTATION_LAMBDA_HANDLER" : "AWS_KINESIS_DATA_STREAM"
    }
  }
}
```

### Conditions (condition/)

#### Conditioner Interface

The `Inspector` interface was renamed to `Conditioner` to standardize the naming convention used across the project.

#### `meta.condition` Condition

This is replaced by the `meta.all`, `meta.any`, and `meta.none` conditions.

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

#### `meta.for_each` Condition

This is replaced by the `meta.all`, `meta.any`, and `meta.none` conditions. If the `object.source_key` value is an array, then the data is treated as a list of elements.

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
sub.cnd.meta.any({
  object: { source_key: 'field' },
  conditions: [ sub.cnd.str.eq({ value: 'FOO' }) ],
})
```

#### `meta.negate` Condition

This is replaced by the `meta.none` Condition.

v1.x.x:

```jsonnet
sub.cnd.meta.negate({ inspector: sub.cnd.str.eq({ value: 'FOO' }) })
```

v2.x.x:

```jsonnet
sub.cnd.meta.none({ conditions: [ sub.cnd.str.eq({ value: 'FOO' }) ] })
```


```jsonnet
sub.cnd.none([ sub.cnd.str.eq({ value: 'FOO' }) ])
```

#### `meta.err` Condition

This is removed and was not replaced. Remove any references to this inspector.

### Transforms (transforms)

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
  arn: 'arn:aws:sqs:us-east-1:123456789012:substation',
  retry: { count: 3 },
})
```

v2.x.x:

```jsonnet
sub.tf.meta.retry({
  retry: { count: 3, delay: '1s' },
  transforms: [
    sub.tf.send.aws.sqs({
      aws: { arn: 'arn:aws:sqs:us-east-1:123456789012:substation' },
    }),
  ],
})
```

**NOTE: For AWS services, retries for the client can be configured in Terraform by using the AWS_MAX_ATTEMPTS environment variable. This is used _in addition_ the `meta.retry` transform.**

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
sub.tf.obj.cp({ object: { src: 'id', trg: 'meta ddb.PK' } }),
sub.transform.enrich.aws.dynamodb({
  object: { source_key: 'meta ddb', target_key: 'user' },
  table_name: 'substation',
  partition_key: 'PK',
  key_condition_expression: 'PK = :PK',
}),
```

v2.x.x:

```jsonnet
sub.transform.enrich.aws.dynamodb.query({
  object: { source_key: 'id', target_key: 'user' },
  aws: { arn: 'arn:aws:dynamodb:us-east-1:123456789012:table/substation' },
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
