local sub = import '../../../../../build/config/substation.libsonnet';

local const = import 'const.libsonnet';
local dynamodb = import 'dynamodb.libsonnet';

{
  sink: sub.interfaces.sink.aws_dynamodb(settings={table:'substation', key:const.ddb_payload}),
  // use the batch transform to modify before it's written to DynamoDB.
  transform: {
    type: 'batch',
    settings: {
      processors:
        dynamodb.processors,
    },
  },
}
