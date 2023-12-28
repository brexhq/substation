local sub = import '../../../../build/config/substation.libsonnet';

local match = sub.cnd.any(
  sub.cnd.string.equal_to({ obj: { src: 'foo' }, string: 'baz' }),
);

local copy = sub.tf.object.copy({ obj: { src: 'foo', dst: 'bar' } },);

{
  transforms: [
    sub.pattern.tf.conditional(
      condition=match, transform=copy,
    ),
    sub.tf.object.insert({ obj: { dst: 'qux' }, value: 'quux' },),
  ],
}
