local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.time.now({ object: { target_key: 'ts' } }),
    sub.tf.obj.insert({ object: { target_key: 'message' }, value: 'Hello from the EventBridge scheduler!' }),
    // This sends the event to the default bus.
    sub.tf.send.aws.eventbridge(),
  ],
}
