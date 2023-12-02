local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // Remove any events that do not have a 'uuid' field.
    sub.pattern.transform.conditional(
      condition=sub.condition.all(sub.pattern.condition.number.length.eq_zero(key='uuid')),
      transform=sub.transform.utility.drop(),
    ),
    // Performs a reverse DNS lookup on the 'addr' field if it is a public IP address.
    sub.pattern.transform.conditional(
      condition=sub.condition.none(sub.pattern.condition.network.ip.internal(key='addr')),
      transform=sub.transform.enrich.dns.ip_lookup(
        settings={ object: { key: 'addr', set_key: 'domain' } },
      ),
    ),
    // The uuid field is used as the partition key for the DynamoDB table.
    sub.transform.object.copy(
      settings={ object: { key: 'uuid', set_key: 'PK' } }
    ),
    sub.transform.object.delete(
      settings={ object: { key: 'uuid' } }
    ),
    sub.transform.send.aws.dynamodb(
      settings={ table_name: 'substation' }
    ),
  ],
}
