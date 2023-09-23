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
    // The SHA256 hash of the event is used as the UUID (Partition Key) for the microservice.
    sub.transform.hash.sha256(
      settings={ object: { key: '@this', set_key: 'PK' } },
    ),
    sub.transform.send.aws.sqs(
      settings={ queue: 'substation' }
    ),
  ],
}
