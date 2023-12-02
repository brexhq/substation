local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.pattern.tf.conditional(
      condition=sub.cnd.all([
        sub.cnd.number.length.greater_than(
          settings={ object: { key: 'body' }, length: 0 }
        ),
      ]),
      transform=sub.tf.object.copy(
        settings={ object: { key: 'body' } }
      ),
    ),
    // Performs a reverse DNS lookup on the 'addr' field if it is a public IP address.
    sub.pattern.tf.conditional(
      condition=sub.cnd.none(sub.pattern.cnd.network.ip.internal(key='addr')),
      transform=sub.tf.enrich.dns.ip_lookup(
        settings={ object: { key: 'addr', set_key: 'domain' } },
      ),
    ),
    // The DNS response is copied so that it is the only value returned in the object.
    sub.tf.object.copy(
      settings={ object: { key: 'domain' } },
    ),
    sub.tf.object.copy(
      settings={ object: { set_key: 'domain' } },
    ),
  ],
}
