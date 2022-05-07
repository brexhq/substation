local conditionlib = import '../../condition.libsonnet';
local sinklib = import '../../sink.libsonnet';

local attributes = [
  {
    key: 'event.hash',
    attribute: 'pk',
  },
  {
    key: 'event.created',
    attribute: 'event_created',
  },
];

{
  sink: sinklib.dynamodb(table='substation_example', attributes=attributes),
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
