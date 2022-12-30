{
  helpers: {
    // dynamically flattens processor configurations
    flatten_processors(processors): std.flattenArrays([
      if std.objectHas(p, 'processors') then p.processors
      else if std.objectHas(p, 'processor') then [p.processor]
      else [p]
      for p in processors
    ]),
  },
  // apply mirrors the applier interface
  apply(options,
        key='',
        set_key='',
        condition={},
        ignore_close=false,
        ignore_errors=false): {
    settings: {
      options: if std.objectHas(options, 'opts') then options.opts else null,
      key: key,
      set_key: set_key,
      condition: condition,
      ignore_close: ignore_close,
      ignore_errors: ignore_errors,
    },
    type: options.type,
  },
  aggregate(key='',
            separator='',
            max_count=1000,
            max_size=10000): {
    type: 'aggregate',
    opts: {
      key: key,
      separator: separator,
      max_count: max_count,
      max_size: max_size,
    },
  },
  aws_dynamodb(table,
               key_condition_expression,
               limit=1,
               scan_index_forward=false): {
    type: 'aws_dynamodb',
    opts: {
      table: table,
      key_condition_expression: key_condition_expression,
      limit: limit,
      scan_index_forward: scan_index_forward,
    },
  },
  aws_lambda(function_name): {
    type: 'aws_lambda',
    opts: {
      function_name: function_name,
    },
  },
  base64(direction): {
    type: 'base64',
    opts: {
      direction: direction,
    },
  },
  capture(expression,
          type='find',
          count=-1): {
    type: 'capture',
    opts: {
      expression: expression,
      type: type,
      count: count,
    },
  },
  case(type): {
    type: 'case',
    opts: {
      type: type,
    },
  },
  convert(type): {
    type: 'convert',
    opts: {
      type: type,
    },
  },
  copy: {
    type: 'copy',
  },
  delete: {
    type: 'delete',
  },
  dns(type,
      timeout=1000): {
    type: 'dns',
    opts: {
      type: type,
      timeout: timeout,
    },
  },
  domain(type): {
    type: 'domain',
    opts: {
      type: type,
    },
  },
  drop: {
    type: 'drop',
  },
  expand: {
    type: 'expand',
  },
  flatten(deep=true): {
    type: 'flatten',
    opts: { deep: deep },
  },
  for_each(processor): {
    type: 'for_each',
    opts: {
      processor: processor,
    },
  },
  group(keys=[]): {
    type: 'group',
    opts: { keys: keys },
  },
  gzip(direction): {
    type: 'gzip',
    opts: { direction: direction },
  },
  hash(algorithm='sha256'): {
    type: 'hash',
    opts: { algorithm: algorithm },
  },
  insert(value): {
    type: 'insert',
    opts: { value: value },
  },
  ip_database(options): {
    type: 'ip_database',
    opts: options,
  },
  join(separator): {
    type: 'join',
    opts: {
      separator: separator,
    },
  },
  math(operation): {
    type: 'math',
    opts: { operation: operation },
  },
  pipeline(processors): {
    type: 'pipeline',
    opts: { processors: processors },
  },
  pretty_print(direction): {
    type: 'pretty_print',
    opts: { direction: direction },
  },
  replace(old,
          new,
          count=-1): {
    type: 'replace',
    opts: { old: old, new: new, count: count },
  },
  split(separator): {
    type: 'split',
    opts: { separator: separator },
  },
  time(format,
       location='',
       set_format='2006-01-02T15:04:05.000000Z',
       set_location=''): {
    type: 'time',
    opts: {
      format: format,
      location: location,
      set_format: set_format,
      set_location: set_location,
    },
  },
}
