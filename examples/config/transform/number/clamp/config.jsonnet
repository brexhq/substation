// This example uses the `number.clamp` pattern to return a value that is
// constrained to a range, where the range is defined by two constants.
local sub = import '../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  // Use `null` for object keys to operate on the entire message.
  transforms: sub.pattern.tf.num.clamp(null, null, 0, 100) +
              [
                sub.tf.send.stdout(),
              ],
}
