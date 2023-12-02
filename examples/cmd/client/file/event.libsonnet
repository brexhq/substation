local sub = import '../../../build/config/substation.libsonnet';

local match = sub.cnd.any(
  sub.cnd.string.equal_to(
    settings={ object: { key: 'foo' }, string: 'baz' }
  ),
);

local copy = sub.tf.object.copy(
  settings={ object: { key: 'foo', set_key: 'bar' } },
);

{
  transforms: [
    sub.pattern.tf.conditional(
      condition=match, transform=copy,
    ),
    sub.tf.object.insert(
      settings={ object: { set_key: 'qux' }, value: 'quux' },
    ),
  ],
}
