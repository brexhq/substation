local sinklib = import '../../../config/../config/sink.libsonnet';

local dynamodb = import './dynamodb.libsonnet';

{
  sink: sinklib.dynamodb(table='substation_example', items_key='__tmp.ddb'),
  transform: {
    type: 'process',
    settings: {
      processors:
        dynamodb.processors
    },
  },
}
