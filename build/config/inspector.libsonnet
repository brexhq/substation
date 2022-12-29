{
  // inspect mirrors the inspector interface
  inspect(options, key='', negate=false): {
    settings: {
      key: key,
      negate: negate,
      options: options.opts,
    },
    type: options.type,
  },
  content(type): {
    type: 'content',
    opts: {
      type: type,
    },
  },
  for_each(type, inspector): {
    type: 'for_each',
    opts: {
      type: type,
      inspector: inspector,
    },
  },
  ip(type): {
    type: 'ip',
    opts: {
      type: type,
    },
  },
  json_schema(schema): {
    type: 'json_schema',
    opts: {
      schema: schema,
    },
  },
  json_valid: {
    type: 'json_valid',
  },
  length(type, value, measurement='bytes'): {
    type: 'length',
    opts: {
      type: type,
      value: value,
      measurement: measurement,
    },
  },
  random: {
    type: 'random',
  },
  regexp(expression): {
    type: 'regexp',
    opts: {
      expression: expression,
    },
  },
  strings(type, expression): {
    type: 'strings',
    opts: {
      type: type,
      expression: expression,
    },
  },
}
