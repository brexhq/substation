local lib = import '../../../../build/config/interfaces.libsonnet';

local consts = import 'consts.libsonnet';
local dynamodb = import 'dynamodb.libsonnet';

{
  sink: lib.sink.aws_dynamodb(table='substation_example', key=consts.ddb_payload),
  // use the batch transform to modify before it's written to DynamoDB.
  transform: {
    type: 'batch',
    settings: {
      processors:
        dynamodb.processors,
    },
  },
}
