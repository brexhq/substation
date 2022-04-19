local conditionlib = import '../../condition.libsonnet';

{
  sink: {
    type: 'dynamodb',
    settings: {
      items: [
        {
          // matches JSON events that have a non-empty `event.hash` key
          condition: {
            operator: 'and',
            inspectors: [
              conditionlib.strings.empty('event.hash', negate=true),
            ],
          },
          // writes `event.hash` values to the pk (partition key) column and `event.created` values to the `event_created` column in the `substation_example` table
          table: 'substation_example',
          fields: [
            {
              key: 'event.hash',
              attribute: 'pk',
            },
            {
              key: 'event.created',
              attribute: 'event_created',
            },
          ],
        },
      ],
    },
  },
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
