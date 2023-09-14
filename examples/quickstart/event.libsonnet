local sub = import '../../build/config/substation.libsonnet';

local match = sub.condition.any(
  sub.condition.string.equal_to(
    settings={ object: { key: 'foo' }, string: 'baz' }
  ),
);

local copy = sub.transform.object.copy(
  settings={ object: { key: 'foo', set_key: 'bar' } },
);

{
  transforms: [
    sub.patterns.transform.conditional(
      condition=match, transform=copy,
    ),
    sub.transform.object.insert(
      settings={ object: { set_key: 'qux' }, value: 'quux' },
    ),
  ],
}
