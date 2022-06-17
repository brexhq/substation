{
  base64(input, output, direction, 
         alphabet='std', 
         condition_operator='', condition_inspectors=[]): {
    type: 'base64',
    settings: {
      input: input, output: output,
      options: { direction: direction, alphabet: alphabet },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  capture(input, output, expression, 
          _function='find', count=-1, 
          condition_operator='', condition_inspectors=[]): {
    type: 'capture',
    settings: {
      input: input, output: output,
      options: { expression: expression, 'function': _function, count: count },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  case(input, output, case, 
       condition_operator='', condition_inspectors=[]): {
    type: 'case',
    settings: {
      input: input, output: output,
      options: { case: case },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  concat(input, output, separator, 
         condition_operator='', condition_inspectors=[]): {
    type: 'concat',
    settings: {
      input: input, output: output,
      options: { separator: separator },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  convert(input, output, type, 
          condition_operator='', condition_inspectors=[]): {
    type: 'convert',
    settings: {
      input: input, output: output,
      options: { type: type },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  copy(input='', output='', 
       condition_operator='', condition_inspectors=[]): {
    type: 'copy',
    settings: {
      input: input, output: output,
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  delete(input, 
         condition_operator='', condition_inspectors=[]): {
    type: 'delete',
    settings: {
      input: input,
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  domain(input, output, _function, 
         condition_operator='', condition_inspectors=[]): {
    type: 'domain',
    settings: {
      input: input, output: output,
      options: { 'function': _function },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  drop(condition_operator='', condition_inspectors=[]): {
    type: 'drop',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  dynamodb(input, output, table, key_condition_expression, 
           limit=1, scan_index_forward=false, 
           condition_operator='', condition_inspectors=[]): {
    type: 'dynamodb',
    settings: {
      input: input, output: output,
      options: { table: table, key_condition_expression: key_condition_expression, limit: limit, scan_index_forward: scan_index_forward },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  expand(input, 
         retain=[], 
         condition_operator='', condition_inspectors=[]): {
    type: 'expand',
    settings: {
      input: input,
      options: { retain: retain },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  flatten(input, output, 
          deep=true, 
          condition_operator='', condition_inspectors=[]): {
    type: 'flatten',
    settings: {
      input: input, output: output,
      options: { deep: deep },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  group(input, output, 
        options_keys=[], 
        condition_operator='', condition_inspectors=[]): {
    type: 'group',
    settings: {
      input: input, output: output,
      options: { keys: options_keys },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  gzip(direction, 
       condition_operator='', condition_inspectors=[]): {
    type: 'gzip',
    settings: {
      options: { direction: direction },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  hash(input, output, 
       algorithm='sha256', 
       condition_operator='', condition_inspectors=[]): {
    type: 'hash',
    settings: {
      input: input, output: output,
      options: { algorithm: algorithm },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  insert(output, value, 
         condition_operator='', condition_inspectors=[]): {
    type: 'insert',
    settings: {
      output: output,
      options: { value: value },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  lambda(input, output, _function, 
         error_on_failure=false, 
         condition_operator='', condition_inspectors=[]): {
    type: 'lambda',
    settings: {
      input: input, output: output,
      options: { 'function': _function, error_on_failure: error_on_failure },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  math(input, output, operation, 
       condition_operator='', condition_inspectors=[]): {
    type: 'math',
    settings: {
      input: input, output: output,
      options: { operation: operation },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  replace(input, output, old, new, 
          count=-1, 
          condition_operator='', condition_inspectors=[]): {
    type: 'replace',
    settings: {
      input: input, output: output,
      options: { old: old, new: new, count: count },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
  time(input, output, input_format, 
       input_location='', output_format='2006-01-02T15:04:05.000000Z', output_location='', 
       condition_operator='', condition_inspectors=[]): {
    type: 'time',
    settings: {
      input: input, output: output,
      options: { input_format: input_format, input_location: input_location, output_format: output_format, output_location: output_location },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
    },
  },
}
