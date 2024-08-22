// This example determines if all values in an array are email addresses
// that have the DNS domain "brex.com". This technique can be used to
// validate or summarize values in an array.
local sub = import '../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.stdout(),
    // In real-world deployments, the match decision is typically used
    // to summarize an array of values. For this example, the decision
    // is represented as a boolean value and printed to stdout.
    sub.tf.meta.switch(
      settings={ cases: [
        {
          condition: sub.cnd.meta.any({ 
            object: { source_key: '@this' },  // Required to interpret the input as a JSON array.
            inspectors: [sub.cnd.str.ends_with(settings={ value: '@brex.com' })] 
          }),
          transforms: [
            sub.tf.obj.insert({ object: { target_key: 'meta result' }, value: true })
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
