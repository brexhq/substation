//  This example shows usage of the 'number.equal_to' and 'number.greater_than' conditions.
local sub = import '../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.num.eq({object: { source_key: 'sourcePort'}, value: 22}),
        transforms: [
          sub.tf.obj.insert({ object: { target_key: 'service'}, value: 'SSH' }),
        ],
      }, 
    ] } ),
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.num.gt({object: { source_key: 'bytes'}, value: 10000}),
        transforms: [
          sub.tf.obj.insert({ object: { target_key: 'severity'}, value: 'high' }),
        ],
      },
    ] }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
