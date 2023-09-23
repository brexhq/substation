local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.patterns.transform.conditional(
      transform=sub.transform.object.copy(
        settings={ object: { key: 'body' } }
      ),
      condition=sub.condition.all([
        sub.condition.logic.len.greater_than(
          settings={ object: { key: 'body' }, length: 0 }
        ),
      ]),
    ),
    // Performs a reverse DNS lookup on the 'addr' field if it is a public IP address.
    sub.patterns.transform.conditional(
      condition=sub.condition.none(sub.patterns.condition.network.ip.internal(key='addr')),
      transform=sub.transform.enrich.dns.ip_lookup(
        settings={ object: { key: 'addr', set_key: 'domain' } },
      ),
    ),
    // The DNS response is copied so that it is the only value returned in the object.
    sub.transform.object.copy(
      settings={ object: { key: 'domain' } },
    ),
    sub.transform.object.copy(
      settings={ object: { set_key: 'domain' } },
    ),
  ],
}
