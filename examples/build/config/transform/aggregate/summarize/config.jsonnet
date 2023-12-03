// This example reduces data by summarizing multiple network events into a single event,
// simulating the behavior of flow records. This technique can be used to reduce
// any JSON data that contains common fields, not just network events.
local sub = import '../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Events are aggregated into arrays based on their client and server fields.
    // The resulting array is put into a new field named "reduce".
    sub.tf.object.copy(
      settings={ object: { key: '[client,server]', set_key: 'meta buffer' } }
    ),
    sub.tf.aggregate.to.array(
      settings={ object: { set_key: 'reduce' }, buffer: { key: 'meta buffer' } }
    ),
    // The "reduce" field is then reduced into a new object that contains:
    // - The last event in the array.
    // - The number of events in the array.
    // - The sum of the "bytes" field of all events in the array.
    sub.tf.object.copy(
      settings={ object: { key: 'reduce|@reverse.0', set_key: 'meta reduce' } }
    ),
    sub.tf.object.copy(
      settings={ object: { key: 'reduce.#', set_key: 'meta reduce.count' } }
    ),
    sub.tf.number.math.add(
      settings={ object: { key: 'reduce.#.bytes', set_key: 'meta reduce.bytes_total' } }
    ),
    sub.tf.object.delete(
      settings={ object: { key: 'meta reduce.bytes' } }
    ),
    // The created object overwrites the original event object and is sent to stdout.
    sub.tf.object.copy(
      settings={ object: { key: 'meta reduce' } }
    ),
    sub.tf.send.stdout(),
  ],
}
