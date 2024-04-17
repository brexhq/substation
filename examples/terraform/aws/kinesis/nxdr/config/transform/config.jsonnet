local sub = import '../../../../../../../build/config/substation.libsonnet';

local threat = import '../threat_enrichment.libsonnet';

{
  concurrency: 2,
  transforms:
    threat.transforms + [
      // At this point more transforms can be added and the events can be sent
      // to an external system.
      sub.tf.send.stdout(),
    ],
}
