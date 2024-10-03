// This example determines if all values in an array are email addresses
// that have the DNS domain "brex.com". This technique can be used to
// validate or summarize values in an array.
local sub = import '../../../substation.libsonnet';

{
  tests: [
    {
      name: 'meta',
      transforms: [
        sub.tf.test.message({ value: ["alice@brex.com","bob@brex.com"] }),
        sub.tf.send.stdout(),
      ],
      // Asserts that the message is equal to 'true'.
      condition: sub.cnd.str.eq({ value: 'true' }),
    }
  ],
  transforms: [
    // In real-world deployments, the match decision is typically used
    // to summarize an array of values. For this example, the decision
    // is represented as a boolean value and printed to stdout.
    sub.tf.meta.switch(
      settings={ cases: [
        {
          condition: sub.cnd.meta.any({
            object: { source_key: '@this' },  // Required to interpret the input as a JSON array.
            conditions: [
              sub.cnd.str.ends_with({ value: 'brex.com' }),
            ],
          }),
          transforms: [
            sub.tf.obj.insert({ object: { target_key: 'meta result' }, value: true }),
          ],
        },
        {
          transforms: [
            sub.tf.obj.insert({ object: { target_key: 'meta result' }, value: false }),
          ],
        },
      ] }
    ),
    sub.tf.obj.cp({ object: { source_key: 'meta result' } }),
    sub.tf.send.stdout(),
  ],
}
