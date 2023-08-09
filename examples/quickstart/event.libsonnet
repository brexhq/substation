local sub = import '../../build/config/substation.libsonnet';

local match = sub.interfaces.condition.oper.any(
  sub.interfaces.condition.insp.strings(
    settings={ key: 'foo', type:'equals', expression: 'baz' },
  ),
);

local copy = sub.interfaces.transform.proc.copy(
  settings={ key: 'foo', set_key: 'bar' },
);

{
  transforms: [
    sub.patterns.transform.conditional(
      condition=match, transform=copy,
    ),
    sub.interfaces.transform.proc.insert(
      settings={ set_key: 'qux', value: 'quux' },
    ),
  ]
}
