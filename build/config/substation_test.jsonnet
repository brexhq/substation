local sub = import 'substation.libsonnet';

local key = 'a';
local set_key = 'b';

local transform = sub.transform.object.copy(settings={ key: key, set_key: set_key });
local inspector = sub.condition.format.json();

{
  helpers: {
    make_array: sub.helpers.make_array(key),
    key: {
      append: sub.helpers.key.append(key, set_key),
      append_array: sub.helpers.key.append_array(key),
      get_element: sub.helpers.key.get_element(key, 1),
    },
  },
  patterns: {
    condition: {
      obj: sub.patterns.condition.obj(key),
      negate: sub.patterns.condition.negate(inspector),
      network: {
        ip: {
          internal: sub.patterns.condition.network.ip.internal(key),
        },
      },
      logic: {
        len: {
          eq_zero: sub.patterns.condition.number.length.eq_zero(key),
          gt_zero: sub.patterns.condition.number.length.gt_zero(key),
        },
      },
      string: {
        eq: sub.patterns.condition.string.equal_to('x', key),
        contains: sub.patterns.condition.string.contains('x', key),
        starts_with: sub.patterns.condition.string.starts_with('x', key),
        ends_with: sub.patterns.condition.string.ends_with('x', key),
      },
    },
    transform: {
      conditional: sub.patterns.transform.conditional(inspector, transform),
    },
  },
}
