local sub = import '../../../../build/config/substation.libsonnet';

local consts = import 'consts.libsonnet';
local dynamodb = import 'dynamodb.libsonnet';

{
  sink: sub.interfaces.sink.aws_dynamodb(settings={table:'substation_example', key:consts.ddb_payload}),
  // use the batch transform to modify before it's written to DynamoDB.
  transform: {
    type: 'batch',
    settings: {
      processors:
        dynamodb.processors,
    },
  },
}
