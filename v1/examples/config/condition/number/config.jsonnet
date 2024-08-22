local sub = import '../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    //  This shows example usage of the 'number.equal_to' and 'number.greater_than' conditions.
    sub.tf.meta.switch(
      settings={
        cases: [
          {
            condition: sub.cnd.num.eq({ obj: { src: 'sourcePort' }, value: 22 }),
            transform: sub.tf.obj.insert({ obj: { trg: 'protocol' }, value: 'SSH' }),
          },
        ],
      }
    ),
    sub.tf.meta.switch(
      settings={ cases: [
        {
          condition: sub.cnd.num.gt({ obj: { src: 'bytes' }, value: 10000 }),
          transform: sub.tf.obj.insert({ obj: { trg: 'severity' }, value: 'high' }),
        },
      ] }
    ),
    sub.tf.send.stdout(),
  ],
}
