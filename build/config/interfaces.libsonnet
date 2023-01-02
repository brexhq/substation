{
  // mirrors interfaces from the condition package
  operator: {
    all(i): { operator: 'all', inspectors: if !std.isArray(i) then [i] else i },
    any(i): { operator: 'any', inspectors: if !std.isArray(i) then [i] else i },
    none(i): { operator: 'none', inspectors: if !std.isArray(i) then [i] else i },
  },
  inspector: {
    inspect(options, key=null, negate=false): {
      settings: {
        key: key,
        negate: negate,
        options: if std.objectHas(options, 'opts') then options.opts else null,
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
  },
  // mirrors interfaces from the process package
  process: {
    apply(options,
          key=null,
          set_key=null,
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
    aggregate(key=null,
              separator=null,
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
    jq(query): {
      type: 'jq',
      opts: { query: query },
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
         location=null,
         set_format='2006-01-02T15:04:05.000000Z',
         set_location=null): {
      type: 'time',
      opts: {
        format: format,
        location: location,
        set_format: set_format,
        set_location: set_location,
      },
    },
  },
  // mirrors interfaces from the internal/sink package
  sink: {
    aws_dynamodb(table, key): {
      type: 'aws_dynamodb',
      settings: { table: table, key: key },
    },
    aws_kinesis(stream, partition=null, partition_key=null, shard_redistribution=false): {
      type: 'aws_kinesis',
      settings: { stream: stream, partition: partition, partition_key: partition_key, shard_redistribution: shard_redistribution },
    },
    aws_kinesis_firehose(stream): {
      type: 'aws_kinesis_firehose',
      settings: { stream: stream },
    },
    aws_s3(bucket, prefix=null, prefix_key=null): {
      type: 'aws_s3',
      settings: { bucket: bucket, prefix: prefix, prefix_key: prefix_key },
    },
    aws_sqs(queue): {
      type: 'aws_sqs',
      settings: { queue: queue },
    },
    grpc(server, timeout=null, certificate=null): {
      type: 'grpc',
      settings: { server: server, timeout: timeout, certificate: certificate },
    },
    http(url, headers=[], headers_key=null): {
      type: 'http',
      settings: { url: url, headers: headers, headers_key: headers_key },
    },
    stdout: {
      type: 'stdout',
    },
    sumologic(url, category=null, category_key=null): {
      type: 'sumologic',
      settings: { url: url, category: category, category_key: category_key },
    },
  },
  // mirrors interfaces from the internal/ip_database/database package
  ip_database: {
    ip2location(database): {
      type: 'ip2location',
      settings: { database: database },
    },
    maxmind_asn(database, language='en'): {
      type: 'maxmind_asn',
      settings: { database: database, language: language },
    },
    maxmind_city(database, language='en'): {
      type: 'maxmind_city',
      settings: { database: database, language: language },
    },
  },
}
