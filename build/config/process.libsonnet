{
  process(options,
          condition={},
          key='',
          set_key='',
          ignore_close=false,
          ignore_errors=false): {
    settings: {
      condition: condition,
      key: key,
      set_key: set_key,
      ignore_close: ignore_close,
      ignore_errors: ignore_errors,
      options: options.opts,
    },
    type: options.type,
  },
  aggregate(key='',
            separator='',
            max_count=1000,
            max_size=10000): {
    type: 'aggregate',
    options: {
      key: key,
      separator: separator,
      max_count: max_count,
      max_size: max_size,
    },
  },
  base64(direction): {
    type: 'base64',
    options: {
      direction: direction,
    },
  },
  capture(expression,
          type='find',
          count=-1): {
    type: 'capture',
    options: {
      expression: expression,
      type: type,
      count: count,
    },
  },
  case(type): {
    type: 'case',
    options: {
      type: type,
    },
  },
  convert(type): {
    type: 'convert',
    settings: {
      options: {
        type: type,
      },
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
    settings: {
      options: {
        type: type,
        timeout: timeout,
      },
    },
  },
  domain(type): {
    type: 'domain',
    settings: {
      options: {
        type: type,
      },
    },
  },
  drop: {
    type: 'drop',
  },
  dynamodb(table,
           key_condition_expression,
           limit=1,
           scan_index_forward=false): {
    type: 'dynamodb',
    settings: {
      options: {
        table: table,
        key_condition_expression: key_condition_expression,
        limit: limit,
        scan_index_forward: scan_index_forward,
      },
    },
  },
  expand: {
    type: 'expand',
  },
  flatten(deep=true): {
    type: 'flatten',
    settings: {
      options: { deep: deep },
    },
  },
  for_each(processor): {
    type: 'for_each',
    settings: {
      options: {
        processor: processor,
      },
    },
  },
  group(keys=[]): {
    type: 'group',
    settings: {
      options: { keys: keys },
    },
  },
  gzip(direction): {
    type: 'gzip',
    settings: {
      options: { direction: direction },
    },
  },
  hash(algorithm='sha256'): {
    type: 'hash',
    settings: {
      options: { algorithm: algorithm },
    },
  },
  insert(value): {
    type: 'insert',
    settings: {
      options: { value: value },
    },
  },
  ip_database(options): {
    type: 'ip_database',
    settings: {
      options: options,
    },
  },
  join(separator): {
    type: 'join',
    options: {
      separator: separator,
    },
  },
  lambda(function_name): {
    type: 'lambda',
    settings: {
      options: { 'function_name': function_name },
    },
  },
  math(operation): {
    type: 'math',
    settings: {
      options: { operation: operation },
    },
  },
  pipeline(processors): {
    type: 'pipeline',
    settings: {
      options: { processors: processors },
    },
  },
  pretty_print(direction): {
    type: 'pretty_print',
    settings: {
      options: { direction: direction },
    },
  },
  replace(old,
          new,
          count=-1): {
    type: 'replace',
    settings: {
      options: { old: old, new: new, count: count },
    },
  },
  split(separator): {
    type: 'split',
    settings: {
      options: {
        separator: separator,
      },
    },
  },
  time(input_format,
       input_location='',
       output_format='2006-01-02T15:04:05.000000Z',
       output_location=''): {
    type: 'time',
    settings: {
      options: {
        input_format: input_format,
        input_location: input_location,
        output_format: output_format,
        output_location: output_location,
      },
    },
  },
}
