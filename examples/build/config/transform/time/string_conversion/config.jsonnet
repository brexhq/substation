// This example shows how to convert time values between string formats.
local sub = import '../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  // Substation uses Go's pattern-based time format (https://gobyexample.com/time-formatting-parsing)
  // to convert time values to and from strings. All time values in the system are in epoch / Unix format
  // with nanosecond precision.
  transforms: [
    // This converts the string value to Unix time. 
    sub.tf.time.from.string({obj: {source_key: 'time', target_key: 'time'}, format: '2006-01-02T15:04:05.000Z'}),
    // This converts the Unix time back to a string.
    sub.tf.time.to.string({obj: {source_key: 'time', target_key: 'time'}, format: '2006-01-02T15:04:05'}),
    sub.tf.send.stdout(),
  ],
}
