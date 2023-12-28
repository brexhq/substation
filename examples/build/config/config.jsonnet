local sub = import '../../../build/config/substation.libsonnet';

{
  // Substation application configs always contain an array named `transforms`.
  transforms: [
    // Each transform function is defined in the `substation` library.
    sub.transform.object.insert({ obj: { dst: 'a' }, value: 'b' }),
    // Transform functions can be conditionally applied using the
    // `meta.switch` function.
    sub.transform.meta.switch({
      cases: [
        {
          condition: sub.condition.any(
            sub.condition.string.equal_to({ object: { src: 'a' }, string: 'z' })
          ),
          transform: sub.transform.object.insert({ obj: { dst: 'c' }, value: 'd' }),
        },
      ],
    }),
    // This is identical to the previous example, but uses a pre-defined
    // pattern and library abbreviations.
    sub.pattern.tf.conditional(
      condition=sub.cnd.str.eq({ obj: { src: 'a' }, string: 'z' }),
      transform=sub.tf.obj.insert({ obj: { dst: 'c' }, value: 'd' }),
    ),
    // Applications usually rely on a `send` transform to send results
    // to a destination. These can be defined anywhere in the config.
    sub.tf.send.stdout(),
  ],
}
