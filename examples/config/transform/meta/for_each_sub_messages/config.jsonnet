// This example shows how to create ephemeral sub-messages based on
// values inside of an array. These sub-messages only exist within
// a for-each loop and are not sent through the pipeline as separate
// messages. If the sub-messages need to be processed as new messages,
// then use the `sub.tf.aggregate.from.array` transform instead.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Copies a value from the message data into metadata -- this is later
    // copied into the ephemeral sub-message from the array.
    sub.tf.object.copy({ object: { source_key: 'foo', target_key: 'meta foo' } }),
    // This array value contains JSON objects, like `[{"bar":"barre"},{"baz":"bazaar"}]`.
    sub.tf.meta.for_each({
      object: { source_key: 'array', target_key: 'array' },
      transform: sub.tf.meta.pipe({ transforms: [
        // The value from earlier is copied into the inner JSON object inside the array:
        //   - `{"bar":"barre","foo":"fooer"}
        //   - `{"baz":"bazaar","foo":"fooer"}
        sub.tf.object.copy({ object: { source_key: 'meta foo', target_key: 'foo' } }),
        // At this point something else would usually be done with this data, like
        // sending it to an external destination. For deeper processing, sending it to
        // a streaming data platform like AWS Kinesis is recommended.
        sub.tf.send.stdout(),
        // The array's value is modified and will be written to the `array` key, so
        // to prevent that modification the changes have to be reversed.
        sub.tf.object.delete({ object: { source_key: 'foo' } }),
      ] }),
    }),
    // The original message is sent to stdout for reference.
    sub.tf.send.stdout(),
  ],
}
