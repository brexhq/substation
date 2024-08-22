local sub = import '../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    //  This shows example usage of the 'string.equal_to' and 'string.greater_than' conditions.
    // The string greater than and less than conditions compare lexographically with another static or target_key value.
    sub.tf.meta.switch(
      settings={ cases: [
        {
          condition: sub.cnd.str.eq({ obj: { src: 'action' }, value: 'ACCEPT' }),
          transform: sub.tf.obj.insert({ obj: { trg: 'action' }, value: 'Allow' }),
        },
      ] }
    ),
    sub.tf.meta.switch(
      settings={ cases: [
        {
          condition: sub.cnd.str.gt({ obj: { src: 'vpcId' }, value: 'vpc-1a2b3c4d' }),
          transform: sub.tf.obj.insert({ obj: { trg: 'priority' }, value: 'high' }),
        },
      ] }
    ),
    sub.tf.send.stdout(),
  ],
}
