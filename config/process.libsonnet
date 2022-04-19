{
  capture(input, output, expression, count=1, condition_operator='', condition_inspectors=[]): {
    type: 'capture',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { expression: expression, count: count },
    },
  },
  case(input, output, case, condition_operator='', condition_inspectors=[]): {
    type: 'case',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: {
        case: case,
      },
    },
  },
  concat(inputs, output, separator, condition_operator='', condition_inspectors=[]): {
    type: 'concat',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: {
        keys: inputs,
      },
      output: { key: output },
      options: {
        separator: separator,
      },
    },
  },
  convert(input, output, type, condition_operator='', condition_inspectors=[]): {
    type: 'convert',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: {
        type: type,
      },
    },
  },
  copy(input, output, condition_operator='', condition_inspectors=[]): {
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
      options: {
        'function': _function,
      },
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
        partition_key_key: parition_key_input,
        sort_key_key: sort_key_input,
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
      options: {
        retain: retain,
      },
      input: { key: input },
    },
  },
  flatten(input, output, deep=true, condition_operator='', condition_inspectors=[]): {
    type: 'flatten',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: {
        deep: true,
      },
    },
  },
  hash(input, output, algorithm='sha256', condition_operator='', condition_inspectors=[]): {
    type: 'hash',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: {
        algorithm: algorithm,
      },
    },
  },
  insert(output, value, condition_operator='', condition_inspectors=[]): {
    type: 'insert',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      output: { key: output },
      options: {
        value: value,
      },
    },
  },
  lambda(payload, output, _function, condition_operator='', condition_inspectors=[]): {
    type: 'lambda',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: {
        payload: payload,
      },
      output: { key: output },
      options: {
        'function': _function,
      },
    },
  },
  math(inputs, output, operation, condition_operator='', condition_inspectors=[]): {
    type: 'math',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: {
        keys: inputs,
      },
      output: { key: output },
      options: {
        'operation': operation,
      },
    },
  },
  replace(input, output, operation, old, new, count=-1, condition_operator='all', condition_inspectors=[]): {
    type: 'replace',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: { old: old, new: new, count: count },
    },
  },
  time(input, output, input_format, output_format='2006-01-02T15:04:05.000000Z', condition_operator='', condition_inspectors=[]): {
    type: 'time',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: { key: input },
      output: { key: output },
      options: {
        input_format: input_format,
        output_format: output_format,
      },
    },
  },
  zip(inputs, output, options_keys=[], condition_operator='', condition_inspectors=[]): {
    type: 'zip',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: {
        keys: inputs,
      },
      output: { key: output },
      options: {
        keys: options_keys,
      },
    },
  },
}
