local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.pattern.tf.conditional(
      condition=sub.cnd.all([
        sub.cnd.number.length.greater_than(
          settings={ object: { source_key: 'body' }, value: 0 }
        ),
      ]),
      transform=sub.tf.object.copy(
        settings={ object: { source_key: 'body' } }
      ),
    ),
    // Performs a reverse DNS lookup on the 'addr' field if it is a public IP address.
    sub.pattern.tf.conditional(
      condition=sub.cnd.none(sub.pattern.cnd.network.ip.internal(key='addr')),
      transform=sub.tf.enrich.dns.ip_lookup(
        settings={ object: { source_key: 'addr', target_key: 'domain' } },
      ),
    ),
    // The DNS response is copied so that it is the only value returned in the object.
    sub.tf.object.copy(
      settings={ object: { source_key: 'domain' } },
    ),
    sub.tf.object.copy(
      settings={ object: { target_key: 'domain' } },
    ),
  ],
}
