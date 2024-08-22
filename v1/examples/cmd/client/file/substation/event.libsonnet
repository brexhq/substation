local sub = import '../../../../../build/config/substation.libsonnet';

local match = sub.cnd.any(
  sub.cnd.string.equal_to({ object: { source_key: 'foo' }, string: 'baz' }),
);

local copy = sub.tf.object.copy({ object: { source_key: 'foo', target_key: 'bar' } },);

{
  transforms: [
    sub.pattern.tf.conditional(
      condition=match, transform=copy,
    ),
    sub.tf.object.insert({ object: { target_key: 'qux' }, value: 'quux' },),
  ],
}
