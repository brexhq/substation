local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Dynamically handles input from either Lambda URL or synchronous invocation.
    sub.patterns.transform.conditional(
      condition=sub.interfaces.condition.oper.all([
        sub.interfaces.condition.insp.length(
          settings={key: 'body', type: 'greater_than', value: 0}
        ),
      ]),
      transform=sub.interfaces.transform.proc.copy(
        settings={ key: 'body' }
      ),
    ),
    // Performs a reverse DNS lookup on the 'addr' field if it is a public IP address.
    sub.patterns.transform.conditional(
      condition=sub.patterns.condition.oper.ip.public(key='addr'),
      transform=sub.interfaces.transform.proc.dns(
        settings={ key: 'addr', set_key: 'domain', type: 'reverse_lookup' }
      ),
    ),
    // The DNS response is copied so that it is the only value returned in the object.
    sub.interfaces.transform.proc.copy(
      settings={ key: 'domain' }
    ),
    sub.interfaces.transform.proc.copy(
      settings={ set_key: 'domain' }
    ),
  ],
}
