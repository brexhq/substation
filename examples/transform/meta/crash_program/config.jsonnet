// This example shows how to intentionally crash a program if a transform
// does not produce an output. This technique can be used to provide strict
// guarantees about the result of data transformations.
local sub = std.extVar('sub');

// `key` is the target of the transform that may not produce an output and is
// checked to determine if the transform was successful.
local key = 'c';

{
  tests: [
    // This test should result in a config error if the program crashed.
    {
      name: 'crash_program',
      transforms: [
        sub.tf.test.message({ value: { a: 'b' } }),
      ],
      condition: sub.cnd.num.len.gt({ value: 0 }),
    },
  ],
  transforms: [
    // This simulates a transform that may not produce an output.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.utility.random(),
        transforms: [
          sub.tf.obj.insert({ object: { target_key: key }, value: true }),
        ],
      },
    ] }),
    // If there is no output from the transform, then an error is thrown to crash the program.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.num.len.eq(settings={ object: { source_key: key }, value: 0 }),
        transforms: [
          sub.tf.util.err(settings={ message: 'transform produced no output' }),
        ],
      },
      {
        transforms: [
          sub.tf.send.stdout(),
        ],
      },
    ] }),
  ],
}
