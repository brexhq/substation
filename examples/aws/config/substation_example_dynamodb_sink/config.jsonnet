local sink = import '../../../../build/config/sink.libsonnet';

local consts = import 'consts.libsonnet';
local dynamodb = import 'dynamodb.libsonnet';

{
  sink: sink.aws_dynamodb(table='substation_example', key=consts.ddb_payload),
  transform: {
    type: 'batch',
    settings: {
      processors:
        dynamodb.processors,
    },
  },
}
