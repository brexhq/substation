local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'default_value',
      transforms: [
        sub.tf.test.message({ value: { a: 'b' } }),
        sub.tf.send.stdout(),
      ],
      // Asserts that 'z' exists.
      condition: sub.cnd.num.len.greater_than({ obj: { src: 'z' }, value: 0 }),
    },
  ],
  transforms: [
    sub.tf.object.insert({ object: { target_key: 'z' }, value: true }),
    sub.tf.send.stdout(),
    // This simulates a transform that may not produce an output.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.utility.random(),
        transforms: [
          sub.tf.obj.insert({ object: { target_key: 'z' }, value: false }),
        ],
      },
    ] }),
    sub.tf.send.stdout(),
  ],
}
