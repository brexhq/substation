local sub = import '../../../../../../../build/config/substation.libsonnet';

local kv = sub.kv_store.aws_dynamodb({
  table_name: 'substation',
  attributes: { partition_key: 'PK', ttl: 'ttl' },
});

{
  transforms: [
    // All messages are locked before they are sent through other
    // transform functions, ensuring that the message is processed
    // exactly once.
    //
    // An error in any sub-transform will cause all previously locked
    // messages to be unlocked; this only applies to messages that have
    // not yet been flushed by a control message. Use the `utility_control`
    // transform to manage how often messages are flushed.
    sub.tf.meta.kv_store.lock(settings={
      kv_store: kv,
      prefix: 'distributed_lock',
      ttl_offset: '1m',
      transform: sub.tf.meta.pipeline({ transforms: [
        // Delaying and simulating an error makes it possible to
        // test message unlocking in real-time (view changes using
        // the DynamoDB console). Uncomment the lines below to see
        // how it works.
        //
        // sub.tf.utility.delay({ duration: '10s' }),
        // sub.pattern.transform.conditional(
        //   condition=sub.cnd.utility.random(),
        //   transform=sub.tf.utility.err({ message: 'simulating error to trigger unlock' }),
        // ),
        //
        // Messages are printed to the console. After this, they are locked
        // and will not be printed again until the lock expires.
        sub.tf.send.stdout(),
      ] }),
    }),
  ],
}
