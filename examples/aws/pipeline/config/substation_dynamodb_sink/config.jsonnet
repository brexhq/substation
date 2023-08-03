local sub = import '../../../../../build/config/substation.libsonnet';

local ddb_payload = '!metadata ddb';

{
  transforms: [
    // Copy the partition key (PK).
    sub.interfaces.transform.proc.copy(
      settings={key:'event.hash', set_key:sub.helpers.key.append(ddb_payload, 'PK')}
    ),
    // Insert extra attributes.
    sub.interfaces.transform.proc.copy(
      settings={key:'event.created', set_key:sub.helpers.key.append(ddb_payload, 'event_created')}
    ),
    // Send to DynamoDB.
    sub.interfaces.transform.send.aws_dynamodb(
      settings={table:'substation', key:ddb_payload}
    )
  ]
}
