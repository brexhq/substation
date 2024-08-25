// This example shows usage of the 'string.equal_to' and 'string.greater_than' conditions.
local sub = import '../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.str.eq({ obj: { src: 'action' }, value: 'ACCEPT' }),
        transforms: [
          // This overwrites the value of the 'action' key.
          sub.tf.obj.insert({ obj: { trg: 'action' }, value: 'Allow' }),
        ],
      },
    ] }),
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.str.gt({ obj: { src: 'vpcId' }, value: 'vpc-1a2b3c4d' }),
        transforms: [
          // This adds a new key-value pair to the object.
          sub.tf.obj.insert({ obj: { trg: 'priority' }, value: 'high' }),
        ],
      },
    ] }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}