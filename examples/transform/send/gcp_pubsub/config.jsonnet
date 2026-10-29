local sub = std.extVar('sub');

{
  transforms: [
    sub.tf.send.gcp.pubsub({
      // Publish the message stream to the Pub/Sub topic `substation-topic`.
      gcp: { resource: 'projects/substation-project/topics/substation-topic' },
      // Message ordering is OFF by default. Ordering keys cap throughput at
      // 1 MB/s per key, so only enable it when strict per-key ordering is
      // required by uncommenting the line below (the value is a message key
      // whose extracted value becomes the ordering key).
      // ordering_key: 'meta group',
    }),
  ],
}
