{
  base64(input, output, direction, alphabet='std', condition_operator='', condition_inspectors=[]): {
    type: 'base64',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { direction: direction, alphabet: alphabet },
    },
  },
  capture(input, output, expression, _function='find', count=-1, condition_operator='', condition_inspectors=[]): {
    type: 'capture',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { expression: expression, 'function': _function, count: count },
    },
  },
  case(input, output, case, condition_operator='', condition_inspectors=[]): {
    type: 'case',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { case: case },
    },
  },
  concat(inputs, output, separator, condition_operator='', condition_inspectors=[]): {
    type: 'concat',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { keys: inputs },
      output: { key: output },
      options: { separator: separator },
    },
  },
  convert(input, output, type, condition_operator='', condition_inspectors=[]): {
    type: 'convert',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { type: type },
    },
  },
  copy(input='', output='', condition_operator='', condition_inspectors=[]): {
    type: 'copy',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
    },
  },
  delete(input, condition_operator='', condition_inspectors=[]): {
    type: 'delete',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
    },
  },
  domain(input, output, _function, condition_operator='', condition_inspectors=[]): {
    type: 'domain',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { 'function': _function },
    },
  },
  drop(condition_operator='', condition_inspectors=[]): {
    type: 'drop',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  dynamodb(parition_key_input, output, table, key_condition_expression, sort_key_input='', limit=1, scan_index_forward=false, condition_operator='', condition_inspectors=[]): {
    type: 'dynamodb',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: {
        partition_key: parition_key_input,
        sort_key: sort_key_input,
      },
      output: { key: output },
      options: {
        table: table,
        key_condition_expression: key_condition_expression,
        limit: limit,
        scan_index_forward: scan_index_forward,
      },
    },
  },
  expand(input, retain=[], condition_operator='', condition_inspectors=[]): {
    type: 'expand',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      options: { retain: retain },
    },
  },
  flatten(input, output, deep=true, condition_operator='', condition_inspectors=[]): {
    type: 'flatten',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { deep: deep },
    },
  },
  group(inputs, output, options_keys=[], condition_operator='', condition_inspectors=[]): {
    type: 'group',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { keys: inputs },
      output: { key: output },
      options: { keys: options_keys },
    },
  },
  gzip(direction, condition_operator='', condition_inspectors=[]): {
    type: 'gzip',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      options: { direction: direction },
    },
  },
  hash(input, output, algorithm='sha256', condition_operator='', condition_inspectors=[]): {
    type: 'hash',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { algorithm: algorithm },
    },
  },
  insert(output, value, condition_operator='', condition_inspectors=[]): {
    type: 'insert',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      output: { key: output },
      options: { value: value },
    },
  },
  lambda(payload, output, _function, error_on_failure=false, condition_operator='', condition_inspectors=[]): {
    type: 'lambda',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { payload: payload },
      output: { key: output },
      options: { 'function': _function, error_on_failure: error_on_failure },
    },
  },
  math(inputs, output, operation, condition_operator='', condition_inspectors=[]): {
    type: 'math',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { keys: inputs },
      output: { key: output },
      options: { operation: operation },
    },
  },
  replace(input, output, old, new, count=-1, condition_operator='', condition_inspectors=[]): {
    type: 'replace',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { old: old, new: new, count: count },
    },
  },
  time(input, output, input_format, input_location='', output_format='2006-01-02T15:04:05.000000Z', output_location='', condition_operator='', condition_inspectors=[]): {
    type: 'time',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: {
        input_format: input_format,
        input_location: input_location,
        output_format: output_format,
        output_location: output_location,
      },
    },
  },
}
