local helpers = {
  // If the input is not an array, then this returns it as an array.
  make_array(i): if !std.isArray(i) then [i] else i,
  abbv(settings): std.mergePatch(settings, {
    object: if std.objectHas(settings, 'object') then $.abbv_obj(settings.object) else if std.objectHas(settings, 'obj') then $.abbv_obj(settings.obj) else null,
    obj: null,
  }),
  abbv_obj(s): {
    source_key: if std.objectHas(s, 'src') then s.src else if std.objectHas(s, 'source_key') then s.source_key else null,
    src: null,
    target_key: if std.objectHas(s, 'trg') then s.trg else if std.objectHas(s, 'target_key') then s.target_key else null,
    trg: null,
    batch_key: if std.objectHas(s, 'btch') then s.batch else if std.objectHas(s, 'batch_key') then s.batch_key else null,
  },
  id(type, settings): std.join('-', [std.md5(type)[:8], std.md5(std.toString(settings))[:8]]),
};

{
  // Mirrors interfaces from the condition package.
  cnd: $.condition,
  condition: {
    all(i): $.condition.meta.all(settings={inspectors: helpers.make_array(i)}),
    any(i): $.condition.meta.any(settings={inspectors: helpers.make_array(i)}),
    none(i): $.condition.meta.none(settings={inspectors: helpers.make_array(i)}),
    meta: {
      all(settings={}): {
        local default = {
          object: $.config.object,
          inspectors: [],
        },

        type: 'all',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      any(settings={}): {
        local default = {
          object: $.config.object,
          inspectors: [],
        },

        type: 'any',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      none(settings={}): {
        local default = {
          object: $.config.object,
          inspectors: [],
        },

        type: 'none',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    str: $.condition.string,
    string: {
      default: {
        object: $.config.object,
        value: null,
      },
      has(settings={}): $.condition.string.contains(settings=settings),
      contains(settings={}): {
        local default = $.condition.string.default,

        type: 'string_contains',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      match(settings={}): {
        local default = {
          object: $.config.object,
          pattern: null,
        },

        type: 'string_match',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
  },
  // Mirrors structs from the internal/config package.
  config: {
    object: { source_key: null, target_key: null, batch_key: null },
  },
  
}
