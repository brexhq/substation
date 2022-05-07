{
  content(type, negate=false): {
    type: 'content',
    settings: { type: type, negate: negate },
  },
  ip: {
    loopback(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'loopback', negate: negate },
    },
    multicast(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'multicast', negate: negate },
    },
    multicast_link_local(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'multicast_link_local', negate: negate },
    },
    private(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'private', negate: negate },
    },
    unicast_global(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'unicast_global', negate: negate },
    },
    unicast_link_local(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'unicast_link_local', negate: negate },
    },
    unspecified(key, negate=false): {
      type: 'ip',
      settings: { key: key, type: 'unspecified', negate: negate },
    },
  },
  regexp(key, expression, negate=false): {
    type: 'regexp',
    settings: { key: key, expression: expression, negate: negate },
  },
  strings: {
    empty(key, negate=true): {
      type: 'strings',
      settings: { key: key, 'function': 'equals', expression: '', negate: negate },
    },
    equals(key, expression, negate=false): {
      type: 'strings',
      settings: { key: key, 'function': 'equals', expression: expression, negate: negate },
    },
    contains(key, expression, negate=false): {
      type: 'strings',
      settings: { key: key, 'function': 'contains', expression: expression, negate: negate },
    },
    endswith(key, expression, negate=false): {
      type: 'strings',
      settings: { key: key, 'function': 'endswith', expression: expression, negate: negate },
    },
    startswith(key, expression, negate=false): {
      type: 'strings',
      settings: { key: key, 'function': 'startswith', expression: expression, negate: negate },
    },
  },
}
