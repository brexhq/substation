local sub = import '../../../../substation.libsonnet';

{
  transforms: [
    sub.tf.object.insert({ object: { target_key: 'z' }, value: true }),
    // This simulates a transform that may not produce an output.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.utility.random(),
        transforms: [
          sub.tf.obj.insert({ object: { target_key: 'z' }, value: false }),
        ],
      },
    ] }),
    sub.tf.object.copy({ source_key: '@pretty' }),
    sub.tf.send.stdout(),
  ],
}
