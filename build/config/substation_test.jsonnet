local sub = import 'substation.libsonnet';

local src = 'a';
local dst = 'b';

local transform = sub.transform.object.copy(settings={ obj: { src: src, dst: dst } });
local inspector = sub.condition.format.json();

{
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
        obj: { src: src, dst: dst },
        pattern: 'a',
        repl: 'b',
      }),
      replace: sub.transform.string.replace({
        object: { src_key: src, dst_key: dst },
        pattern: 'a',
        replacement: 'b',
      }),
      split: sub.transform.string.split({
        object: { src_key: src, dst_key: dst },
        sep: '.',
      }),
    },
  },
  helpers: {
    make_array: sub.helpers.make_array(src),
    key: {
      append: sub.helpers.object.append(src, dst),
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
