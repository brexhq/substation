// This example reduces data by summarizing multiple network events into a single event,
// simulating the behavior of flow records. This technique can be used to reduce
// any JSON data that contains common fields, not just network events.
local sub = import '../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Events are aggregated into arrays based on their client and server fields.
    // The resulting array is put into a new field named "reduce".
    sub.tf.object.copy({ obj: { src: '[client,server]', dst: 'meta buffer' } }),
    sub.tf.aggregate.to.array({ obj: { dst: 'reduce' }, buffer: { key: 'meta buffer' } }),
    // The "reduce" field is then reduced into a new object that contains:
    // - The last event in the array.
    // - The number of events in the array.
    // - The sum of the "bytes" field of all events in the array.
    sub.tf.object.copy({ obj: { src: 'reduce|@reverse.0', dst: 'meta reduce' } }),
    sub.tf.object.copy({ obj: { src: 'reduce.#', dst: 'meta reduce.count' } }),
    sub.tf.number.math.add({ obj: { src: 'reduce.#.bytes', dst: 'meta reduce.bytes_total' } }),
    sub.tf.object.delete({ obj: { src: 'meta reduce.bytes' } }),
    // The created object overwrites the original event object and is sent to stdout.
    sub.tf.object.copy({ obj: { src: 'meta reduce' } }),
    sub.tf.send.stdout(),
  ],
}
