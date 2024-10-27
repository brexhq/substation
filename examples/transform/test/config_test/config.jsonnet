local sub = std.extVar('sub');

{
  tests: [
    {
      // Every test should have a unique name.
      name: 'my_passing_test',
      // Generates the test message '{"a": true}' which
      // is run through the configured transforms and
      // then checked against the condition.
      transforms: [
        sub.tf.test.message({ value: { a: true } }),
      ],
      // Checks if key 'x' == 'true'.
      condition: sub.cnd.all([
        sub.cnd.str.eq({ object: { source_key: 'x' }, value: 'true' }),
      ]),
    },
  ],
  // These transforms process the test message and the result
  // is checked against the condition.
  transforms: [
    // Copies the value of key 'a' to key 'x'.
    sub.tf.obj.cp({ object: { source_key: 'a', target_key: 'x' } }),
  ],
}
