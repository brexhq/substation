local sub = import '../../../../../build/config/substation.libsonnet';

local ddb_payload = '!metadata ddb';

{
  transforms: [
    // Copy the partition key (PK).
    sub.transform.object.copy(
      settings={key:'event.hash', set_key:sub.helpers.key.append(ddb_payload, 'PK')}
    ),
    // Insert extra attributes.
    sub.transform.object.copy(
      settings={key:'event.created', set_key:sub.helpers.key.append(ddb_payload, 'event_created')}
    ),
    // Send to DynamoDB.
    sub.transform.send.aws.dynamodb(
      settings={table:'substation', object: { key: ddb_payload } }
    )
  ]
}
