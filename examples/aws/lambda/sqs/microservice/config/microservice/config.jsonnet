local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  metrics: {
    type: 'aws_cloudwatch_embedded_metrics',
  },
  transforms: [
    // Remove any events that do not have a 'PK' field.
    sub.patterns.transform.conditional(
      condition=sub.condition.all(sub.patterns.condition.logic.len.eq_zero(key='PK')),
      transform=sub.transform.utility.drop(),
    ),
    // Performs a reverse DNS lookup on the 'addr' field if it is a public IP address.
    sub.patterns.transform.conditional(
      condition=sub.condition.none(sub.patterns.condition.network.ip.internal(key='addr')),
      transform=sub.transform.enrich.dns.ip_lookup(
        settings={ object: { key: 'addr', set_key: 'domain' } },
      ),
    ),
    sub.transform.send.aws.dynamodb(
      settings={ table: 'substation' }
    ),
  ],
}
