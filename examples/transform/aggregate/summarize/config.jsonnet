// This example reduces data by summarizing multiple network events into a single event,
// simulating the behavior of flow records. This technique can be used to reduce
// any JSON data that contains common fields, not just network events.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'summarize',
      transforms: [
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.8.8","bytes":11,"timestamp":1674429049} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.4.4","bytes":20,"timestamp":1674429050} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.4.4","bytes":15,"timestamp":1674429051} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.8.8","bytes":8,"timestamp":1674429052} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.8.8","bytes":25,"timestamp":1674429053} }),
        sub.tf.test.message({ value: {"client":"10.1.1.4","server":"1.2.3.4","bytes":2400,"timestamp":1674429054} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.4.4","bytes":23,"timestamp":1674429055} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.8.8","bytes":12,"timestamp":1674429056} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.4.4","bytes":18,"timestamp":1674429057} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.8.8","bytes":6,"timestamp":1674429058} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.4.4","bytes":23,"timestamp":1674429059} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.8.8","bytes":12,"timestamp":1674429060} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.4.4","bytes":18,"timestamp":1674429061} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.8.8","bytes":6,"timestamp":1674429062} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.8.8","bytes":11,"timestamp":1674429063} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.4.4","bytes":20,"timestamp":1674429064} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.4.4","bytes":15,"timestamp":1674429065} }),
        sub.tf.test.message({ value: {"client":"10.1.1.3","server":"8.8.8.8","bytes":8,"timestamp":1674429066} }),
        sub.tf.test.message({ value: {"client":"10.1.1.2","server":"8.8.8.8","bytes":25,"timestamp":1674429067} }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each message has the 'count' and 'bytes_total' fields.
      condition: sub.cnd.all([
        sub.cnd.num.len.greater_than({ obj: {src: 'count'}, value: 0 }),
        sub.cnd.num.len.greater_than({ obj: {src: 'bytes_total'}, value: 0 }),
      ])
    }
  ],
  transforms: [
    // Events are aggregated into arrays based on their client and server fields.
    // The resulting array is put into a new field named "reduce".
    sub.tf.object.copy({ object: { source_key: '[client,server]', target_key: 'meta buffer' } }),
    sub.tf.aggregate.to.array({ object: { target_key: 'reduce', batch_key: 'meta buffer' } }),
    // The "reduce" field is then reduced into a new object that contains:
    // - The last event in the array.
    // - The number of events in the array.
    // - The sum of the "bytes" field of all events in the array.
    sub.tf.object.copy({ object: { source_key: 'reduce|@reverse.0', target_key: 'meta reduce' } }),
    sub.tf.object.copy({ object: { source_key: 'reduce.#', target_key: 'meta reduce.count' } }),
    sub.tf.number.math.add({ object: { source_key: 'reduce.#.bytes', target_key: 'meta reduce.bytes_total' } }),
    sub.tf.object.delete({ object: { source_key: 'meta reduce.bytes' } }),
    // The created object overwrites the original event object and is sent to stdout.
    sub.tf.object.copy({ object: { source_key: 'meta reduce' } }),
    sub.tf.send.stdout(),
  ],
}
