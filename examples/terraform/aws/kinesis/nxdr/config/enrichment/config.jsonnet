local sub = import '../../../../../../../build/config/substation.libsonnet';

local const = import '../const.libsonnet';
local threat = import '../threat_enrichment.libsonnet';

{
  concurrency: 2,
  transforms:
    threat.transforms + [
      // Discards any events that don't contain threat signals.
      sub.tf.meta.switch({ cases: [
        {
          condition: sub.cnd.any([
            sub.cnd.num.len.eq({ object: { source_key: const.threat_signals_key }, value: 0 }),
          ]),
          transform: sub.tf.util.drop(),
        },
      ] }),
      // Explodes the threat signals array into individual events. These become
      // threat signal records in the DynamoDB table.
      sub.tf.aggregate.from.array({ object: { source_key: const.threat_signals_key } }),
      // The host name and current time are used as the keys for the DynamoDB table.
      sub.tf.object.copy({ object: { source_key: 'host.name', target_key: 'PK' } }),
      sub.tf.time.now({ object: { target_key: 'SK' } }),
      sub.tf.time.to.string({ object: { source_key: 'SK', target_key: 'SK' }, format: '2006-01-02T15:04:05.000Z' }),
      // Any fields not needed in the DynamoDB item are removed.
      sub.tf.object.delete({ object: { source_key: 'event' } }),
      sub.tf.object.delete({ object: { source_key: 'host' } }),
      sub.tf.object.delete({ object: { source_key: 'process' } }),
      sub.tf.object.delete({ object: { source_key: 'threat' } }),
      // Writes the threat signal to the DynamoDB table.
      sub.tf.send.aws.dynamodb({ table_name: 'substation_threat_signals' }),
    ],
}
