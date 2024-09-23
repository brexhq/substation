local sub = import 'substation.libsonnet';

local src = 'source';
local trg = 'target';

{
  condition: {
    all: sub.condition.all([$.condition.string.contains, $.condition.string.match]),
    any: sub.condition.any([$.condition.string.contains, $.condition.string.match]),
    none: sub.condition.none([$.condition.string.contains, $.condition.string.match]),
    meta: {
      all: sub.condition.meta.all({ inspectors: [$.condition.string.contains, $.condition.string.match] }),
      any: sub.condition.meta.any({ inspectors: [$.condition.string.contains, $.condition.string.match] }),
      none: sub.condition.meta.none({ inspectors: [$.condition.string.contains, $.condition.string.match] }),
    },
    string: {
      contains: sub.condition.string.contains({ obj: { src: src }, value: 'z' }),
      match: sub.condition.string.match({ obj: { src: src }, pattern: 'z' }),
    },
  },
}
