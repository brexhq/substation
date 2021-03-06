{
  base64(input, output, direction, 
         condition_operator='', condition_inspectors=[]): {
    type: 'base64',
    settings: {
      options: { direction: direction },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  capture(input, output, expression, 
          _function='find', count=-1, 
          condition_operator='', condition_inspectors=[]): {
    type: 'capture',
    settings: {
      options: { expression: expression, 'function': _function, count: count },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  case(input, output, case, 
       condition_operator='', condition_inspectors=[]): {
    type: 'case',
    settings: {
      options: { case: case },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  concat(input, output, separator, 
         condition_operator='', condition_inspectors=[]): {
    type: 'concat',
    settings: {
      options: { separator: separator },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  convert(input, output, type, 
          condition_operator='', condition_inspectors=[]): {
    type: 'convert',
    settings: {
      options: { type: type },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  copy(input='', output='', 
       condition_operator='', condition_inspectors=[]): {
    type: 'copy',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  delete(input, 
         condition_operator='', condition_inspectors=[]): {
    type: 'delete',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: input,
    },
  },
  domain(input, output, _function, 
         condition_operator='', condition_inspectors=[]): {
    type: 'domain',
    settings: {
      options: { 'function': _function },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
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
      options: { table: table, key_condition_expression: key_condition_expression, limit: limit, scan_index_forward: scan_index_forward },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  expand(input, 
         condition_operator='', condition_inspectors=[]): {
    type: 'expand',
    settings: {
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input: input,
    },
  },
  flatten(input, output, 
          deep=true, 
          condition_operator='', condition_inspectors=[]): {
    type: 'flatten',
    settings: {
      options: { deep: deep },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  group(input, output, 
        keys=[], 
        condition_operator='', condition_inspectors=[]): {
    type: 'group',
    settings: {
      options: { keys: keys },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
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
      options: { algorithm: algorithm },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  insert(output, value, 
         condition_operator='', condition_inspectors=[]): {
    type: 'insert',
    settings: {
      options: { value: value },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      output: output,
    },
  },
  lambda(input, output, _function, 
         error_on_failure=false, 
         condition_operator='', condition_inspectors=[]): {
    type: 'lambda',
    settings: {
      options: { 'function': _function, error_on_failure: error_on_failure },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  math(input, output, operation, 
       condition_operator='', condition_inspectors=[]): {
    type: 'math',
    settings: {
      options: { operation: operation },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  replace(input, output, old, new, 
          count=-1, 
          condition_operator='', condition_inspectors=[]): {
    type: 'replace',
    settings: {
      options: { old: old, new: new, count: count },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
  time(input, output, input_format, 
       input_location='', output_format='2006-01-02T15:04:05.000000Z', output_location='', 
       condition_operator='', condition_inspectors=[]): {
    type: 'time',
    settings: {
      options: { input_format: input_format, input_location: input_location, output_format: output_format, output_location: output_location },
      condition: { operator: condition_operator, inspectors: condition_inspectors},
      input_key: input, output_key: output,
    },
  },
}
