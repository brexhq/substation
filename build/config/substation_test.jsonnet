local sub = import 'substation.libsonnet';

local src = 'a';
local trg = 'b';

local transform = sub.transform.object.copy(settings={ obj: { src: src, trg: trg } });
local inspector = sub.condition.format.json();

{
  condition: {
    number: {
      equal_to: sub.condition.number.equal_to({obj: {src: src}, value: 1}),
      less_than: sub.condition.number.less_than({obj: {src: src}, value: 1}),
      greater_than: sub.condition.number.greater_than({obj: {src: src}, value: 1}),
    },
  },
  transform: {
    send: {
      http: {
        post: sub.transform.send.http.post({
          url: 'http://localhost:8080',
          hdr: [{ key: 'Content-Type', value: 'application/json' }],
        }),
      },
    },
    string: {
      repl: sub.transform.string.repl({
        obj: { src: src, trg: trg },
        pattern: 'a',
        repl: 'b',
      }),
      replace: sub.transform.string.replace({
        object: { source_key: src, target_key: trg },
        pattern: 'a',
        replacement: 'b',
      }),
      split: sub.transform.string.split({
        object: { source_key: src, target_key: trg },
        sep: '.',
      }),
    },
  },
  helpers: {
    make_array: sub.helpers.make_array(src),
    key: {
      append: sub.helpers.object.append(src, trg),
      append_array: sub.helpers.object.append_array(src),
      get_element: sub.helpers.object.get_element(src, 1),
    },
  },
  pattern: {
    condition: {
      obj: sub.pattern.condition.obj(src),
      negate: sub.pattern.condition.negate(inspector),
      network: {
        ip: {
          internal: sub.pattern.condition.network.ip.internal(src),
        },
      },
      logic: {
        len: {
          eq_zero: sub.pattern.condition.number.length.eq_zero(src),
          gt_zero: sub.pattern.condition.number.length.gt_zero(src),
        },
      },
    },
    transform: {
      conditional: sub.pattern.transform.conditional(inspector, transform),
    },
  },
}
