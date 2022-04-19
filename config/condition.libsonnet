{
  ip: {
    loopback(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'loopback', negate: negate },
    },
    multicast(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'multicast', negate: negate },
    },
    multicast_link_local(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'multicast_link_local', negate: negate },
    },
    private(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'private', negate: negate },
    },
    unicast_global(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'unicast_global', negate: negate },
    },
    unicast_link_local(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'unicast_link_local', negate: negate },
    },
    unspecified(key, negate=false): {
      type: 'ip',
      settings: { key: key, 'function': 'unspecified', negate: negate },
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
