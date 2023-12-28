// This example shows how to intentionally crash a program if a transform
// does not produce an output. This technique can be used to provide strict
// guarantees about the result of data transformations.
local sub = import '../../../../../../build/config/substation.libsonnet';

// `key` is the target of the transform that may not produce an output and is
// checked to determine if the transform was successful.
local key = 'c';

{
  transforms: [
    // This conditional transform simulates a transform that may not produce an output.
    sub.pattern.tf.conditional(
      condition=sub.cnd.any(sub.cnd.utility.random()),
      transform=sub.tf.obj.insert(settings={ obj: { dst: key }, value: true }),
    ),
    // If there is no output from the transform, then an error is thrown to crash the program.
    sub.tf.meta.switch(settings={ switch: [
      {
        condition: sub.cnd.any(sub.cnd.num.len.eq(settings={ obj: { src: key }, value: 0 })),
        transform: sub.tf.util.err(settings={ message: 'transform produced no output' }),
      },
      { transform: sub.tf.send.stdout() },
    ] }),
  ],
}
