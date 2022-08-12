local sinklib = import '../../../../config/sink.libsonnet';

local dynamodb = import './dynamodb.libsonnet';

{
  sink: sinklib.dynamodb(table='substation_example', items_key='__tmp.ddb'),
  transform: {
    type: 'batch',
    settings: {
      processors:
        dynamodb.processors
    },
  },
}
