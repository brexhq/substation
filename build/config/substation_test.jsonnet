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
  pattern: {
    condition: {
      obj: sub.pattern.condition.obj(key),
      negate: sub.pattern.condition.negate(inspector),
      network: {
        ip: {
          internal: sub.pattern.condition.network.ip.internal(key),
        },
      },
      logic: {
        len: {
          eq_zero: sub.pattern.condition.number.length.eq_zero(key),
          gt_zero: sub.pattern.condition.number.length.gt_zero(key),
        },
      },
    },
    transform: {
      conditional: sub.pattern.transform.conditional(inspector, transform),
    },
  },
}
