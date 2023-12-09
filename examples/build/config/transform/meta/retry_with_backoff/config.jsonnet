// This example shows how to implement retry with backoff behavior for any
// transform that does not produce an output. This technique may be useful
// when enriching data with external services or asynchronous data pipelines.
local sub = import '../../../../../../build/config/substation.libsonnet';

// `key` is the target of the transform that may not produce an output and is
// checked to determine if the transform was successful.
local key = 'c';
local key_is_empty = sub.cnd.num.len.eq(settings={ object: { key: key }, length: 0 });

local cnd = sub.cnd.all([
  key_is_empty,
  // Randomness simulates a transform that may fail to produce an output.
  sub.cnd.utility.random(),
]);

// The number of retries and the backoff duration can be customized. This will
// retry up to 3 times with a backoff duration of 1 second, 2 seconds, and 4 seconds.
local retries = ['0s', '1s', '2s', '4s'];
{
  transforms:
    // The retry with backoff behavior is implemented by pipelining a delay transform
    // with another transform and validating the Message before each retry. The delay
    // duration is increased with each attempt.
    [
      sub.pattern.tf.conditional(
        condition=cnd,
        transform=sub.tf.meta.pipe(settings={ transforms: [
          sub.tf.util.delay(settings={ duration: r }),
          sub.tf.obj.insert(settings={ obj: { set_key: key }, value: true }),
          // This is added to show the number of retries that were attempted. If
          // needed in real-world deployments, then it's recommended to put this
          // info into the Message metadata.
          sub.tf.obj.insert(settings={ obj: { set_key: 'retries' }, value: std.find(r, retries)[0] }),
        ] }),
      )

      for r in retries
    ] + [
      // If there is no output after all retry attempts, then an error is thrown to crash the program.
      // This is the same technique from the build/config/transform/meta/crash_program example.
      sub.tf.meta.switch(settings={ switch: [
        {
          condition: sub.cnd.any(key_is_empty),
          transform: sub.tf.util.err(settings={
            message: std.format('failed to transform after retrying %d times', std.length(retries) - 1),
          }),
        },
        { transform: sub.tf.send.stdout() },
      ] }),
    ],
}
