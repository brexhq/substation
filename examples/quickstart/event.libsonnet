local sub = import '../../build/config/substation.libsonnet';

local match = sub.condition.oper.any(
  sub.condition.insp.string(
    settings={ key: 'foo', type:'equals', expression: 'baz' },
  ),
);

local copy = sub.transform.proc.copy(
  settings={ key: 'foo', set_key: 'bar' },
);

{
  transforms: [
    sub.patterns.transform.conditional(
      condition=match, transform=copy,
    ),
    sub.transform.proc.insert(
      settings={ set_key: 'qux', value: 'quux' },
    ),
  ]
}
