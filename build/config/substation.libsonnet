{
  helpers: {
    // if input is not an array, then this returns an array
    make_array(i): if !std.isArray(i) then [i] else i,
    key: {
      // if key is foo and arr is bar, then result is foo.bar
      // if key is foo and arr is [bar, baz], then result is foo.bar.baz
      append(key, arr): std.join('.', $.helpers.make_array(key) + $.helpers.make_array(arr)),
      // if key is foo, then result is foo.-1
      append_array(key): key + '.-1',
      // if key is foo and e is 0, then result is foo.0
      get_element(key, e=0): std.join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
    },
  },
  defaults: {
    condition: {
      insp: {
        content: {
          settings: { key: null, negate: null, type: null },
        },
        ip: {
          settings: { key: null, negate: null, type: null },
        },
        length: {
          settings: { key: null, negate: null, type: null, length: null, measurement: 'byte' },
        },
        regexp: {
          settings: { key: null, negate: null, expression: null },
        },
        string: {
          settings: { key: null, negate: null, type: null, string: null },
        },
      },
      meta: {
        condition: {
          settings: { negate: null, condition: null },
        },
        for_each: {
          settings: { key: null, negate: null, type: null, inspector: null },
        },
      },
    },
    transform: {
      meta: {
        for_each: {
          settings: { key: null, set_key: null, transform: null },
        },
        pipeline: {
          settings: { key: null, set_key: null, transforms: null },
        },
        switch: {
          settings: { switch: null },
        },
      },
      proc: {
        aws_dynamodb: {
          settings: { key: null, set_key: null, table: null, key_condition_expression: null, limit: 1, scan_index_forward: false },
        },
        aws_lambda: {
          settings: { key: null, set_key: null, function_name: null },
        },
        base64: {
          settings: { key: null, set_key: null, direction: null },
        },
        capture: {
          settings: { key: null, set_key: null, expression: null, type: 'find', count: -1 },
        },
        case: {
          settings: { key: null, set_key: null, type: null },
        },
        combine: {
          settings: { key: null, set_key: null, combine_key: null, separator: null, max_count: 1000, max_size: 10000 },
        },
        convert: {
          settings: { key: null, set_key: null, type: null },
        },
        copy: {
          settings: { key: null, set_key: null },
        },
        delete: {
          settings: {},
        },
        dns: {
          settings: { key: null, set_key: null, type: null, timeout: 1000 },
        },
        domain: {
          settings: { key: null, set_key: null, type: null },
        },
        err: {
          settings: { err: null },
        },
        flatten_array: {
          settings: { key: null, set_key: null, deep: true },
        },
        group: {
          settings: { key: null, set_key: null, keys: null },
        },
        gzip: {
          settings: { key: null, set_key: null, direction: null },
        },
        hash: {
          settings: { key: null, set_key: null, algorithm: 'sha256' },
        },
        http: {
          settings: { key: null, set_key: null, method: 'get', url: null, headers: null, body_key: null },
        },
        insert: {
          settings: { set_key: null, value: null },
        },
        join: {
          settings: { key: null, set_key: null, separator: null },
        },
        jq: {
          settings: { query: null },
        },
        kv_store: {
          settings: { key: null, set_key: null, type: null, prefix: null, ttl_key: null, ttl_offset: null, kv_options: null },
        },
        math: {
          settings: { key: null, set_key: null, operation: null },
        },
        pretty_print: {
          settings: { direction: null },
        },
        replace: {
          settings: { key: null, set_key: null, old: null, new: null, count: -1 },
        },
        split: {
          settings: { key: null, set_key: null, separator: null },
        },
        time: {
          settings: { key: null, set_key: null, format: null, location: null, set_format: '2006-01-02T15:04:05.000000Z', set_location: null },
        },
      },
      send: {
        aws_dynamodb: {
          settings: { table: null, key: null },
        },
        aws_kinesis: {
          settings: { stream: null, partition: null, partition_key: null, shard_redistribution: false },
        },
        aws_kinesis_firehose: {
          settings: { stream: null },
        },
        aws_s3: {
          settings: { bucket: null, file_path: null, file_format: null, file_compression: null },
        },
        aws_sns: {
          settings: { arn: null },
        },
        aws_sqs: {
          settings: { queue: null },
        },
        file: {
          settings: { file_path: null, file_format: { type: 'json' }, file_compression: { type: 'gzip' } },
        },
        http: {
          settings: { url: null, headers: null, headers_key: null },
        },
        stdout: {
          settings: null,
        },
        sumologic: {
          settings: { url: null, category: null, category_key: null },
        },
      },
    },
    kv_store: {
      aws_dynamodb: {
        settings: { table: null, attributes: { partition_key: null, sort_key: null, value: null, ttl: null } },
      },
      csv_file: {
        settings: { file: null, column: null, delimiter: ',', header: null },
      },
      json_file: {
        settings: { file: null, is_lines: false },
      },
      memory: {
        settings: { capacity: 1024 },
      },
      mmdb: {
        settings: { file: null },
      },
      text_file: {
        settings: { file: null },
      },
    },
  },
  interfaces: {
    // Mirrors interfaces from the condition package.
    condition: {
      oper: {
        all(i): { operator: 'all', inspectors: if !std.isArray(i) then [i] else i },
        any(i): { operator: 'any', inspectors: if !std.isArray(i) then [i] else i },
        none(i): { operator: 'none', inspectors: if !std.isArray(i) then [i] else i },
      },
      meta: {
        condition(settings=$.defaults.condition.meta.condition.settings): {
          type: 'meta_condition',
          settings: std.mergePatch($.defaults.condition.meta.condition.settings, settings),
        },
        for_each(settings=$.defaults.condition.meta.for_each.settings): {
          type: 'meta_for_each',
          settings: std.mergePatch($.defaults.condition.meta.for_each.settings, settings),
        },
      },
      insp: {
        content(settings=$.defaults.condition.insp.content.settings): {
          type: 'insp_content',
          settings: std.mergePatch($.defaults.condition.insp.content.settings, settings),
        },
        ip(settings=$.defaults.condition.insp.ip.settings): {
          type: 'insp_ip',
          settings: std.mergePatch($.defaults.condition.insp.ip.settings, settings),
        },
        json_valid(settings=$.defaults.condition.insp.json_valid.settings): {
          type: 'insp_json_valid',
          settings: std.mergePatch($.defaults.condition.insp.json_valid.settings, settings),
        },
        length(settings=$.defaults.condition.insp.length.settings): {
          type: 'insp_length',
          settings: std.mergePatch($.defaults.condition.insp.length.settings, settings),
        },
        random: {
          type: 'insp_random',
        },
        regexp(settings=$.defaults.condition.insp.regexp.settings): {
          type: 'insp_regexp',
          settings: std.mergePatch($.defaults.condition.insp.regexp.settings, settings),
        },
        string(settings=$.defaults.condition.insp.string.settings): {
          type: 'insp_string',
          settings: std.mergePatch($.defaults.condition.insp.string.settings, settings),
        },
      },
    },
    // Mirrors interfaces from the transform package.
    transform: {
      meta: {
        for_each(settings=$.defaults.transform.meta.for_each.settings): {
          type: 'meta_for_each',
          settings: std.mergePatch($.defaults.transform.meta.for_each.settings, settings),
        },
        pipeline(settings=$.defaults.transform.meta.pipeline.settings): {
          type: 'meta_pipeline',
          settings: std.mergePatch($.defaults.transform.meta.pipeline.settings, settings),
        },
        switch(settings=$.defaults.transform.meta.switch.settings): {
          type: 'meta_switch',
          settings: std.mergePatch($.defaults.transform.meta.switch.settings, settings),
        },
      },
      proc: {
        aws_dynamodb(settings=$.defaults.transform.proc.aws_dynamodb.settings): {
          type: 'proc_aws_dynamodb',
          settings: std.mergePatch($.defaults.transform.proc.aws_dynamodb.settings, settings),
        },
        aws_lambda(settings=$.defaults.transform.proc.aws_lambda.settings): {
          type: 'proc_aws_lambda',
          settings: std.mergePatch($.defaults.transform.proc.aws_lambda.settings, settings),
        },
        base64(settings=$.defaults.transform.proc.base64.settings): {
          type: 'proc_base64',
          settings: std.mergePatch($.defaults.transform.proc.base64.settings, settings),
        },
        capture(settings=$.defaults.transform.proc.capture.settings): {
          type: 'proc_capture',
          settings: std.mergePatch($.defaults.transform.proc.capture.settings, settings),
        },
        case(settings=$.defaults.transform.proc.case.settings): {
          type: 'proc_case',
          settings: std.mergePatch($.defaults.transform.proc.case.settings, settings),
        },
        combine(settings=$.defaults.transform.proc.combine.settings): {
          type: 'proc_combine',
          settings: std.mergePatch($.defaults.transform.proc.combine.settings, settings),
        },
        convert(settings=$.defaults.transform.proc.convert.settings): {
          type: 'proc_convert',
          settings: std.mergePatch($.defaults.transform.proc.convert.settings, settings),
        },
        copy(settings=$.defaults.transform.proc.copy.settings): {
          type: 'proc_copy',
          settings: std.mergePatch($.defaults.transform.proc.copy.settings, settings),
        },
        delete(settings=$.defaults.transform.proc.delete.settings): {
          type: 'proc_delete',
          settings: std.mergePatch($.defaults.transform.proc.delete.settings, settings),
        },
        dns(settings=$.defaults.transform.proc.dns.settings): {
          type: 'proc_dns',
          settings: std.mergePatch($.defaults.transform.proc.dns.settings, settings),
        },
        domain(settings=$.defaults.transform.proc.domain.settings): {
          type: 'proc_domain',
          settings: std.mergePatch($.defaults.transform.proc.domain.settings, settings),
        },
        drop(settings=$.defaults.transform.proc.drop.settings): {
          type: 'proc_drop',
          settings: std.mergePatch($.defaults.transform.proc.drop.settings, settings),
        },
        err(settings=$.defaults.transform.proc.err.settings): {
          type: 'proc_err',
          settings: std.mergePatch($.defaults.transform.proc.err.settings, settings),
        },
        expand(settings=$.defaults.transform.proc.expand.settings): {
          type: 'proc_expand',
          settings: std.mergePatch($.defaults.transform.proc.expand.settings, settings),
        },
        flatten_array(settings=$.defaults.transform.proc.flatten_array.settings): {
          type: 'proc_flatten_array',
          settings: std.mergePatch($.defaults.transform.proc.flatten_array.settings, settings),
        },
        group(settings=$.defaults.transform.proc.group.settings): {
          type: 'proc_group',
          settings: std.mergePatch($.defaults.transform.proc.group.settings, settings),
        },
        gzip(settings=$.defaults.transform.proc.gzip.settings): {
          type: 'proc_gzip',
          settings: std.mergePatch($.defaults.transform.proc.gzip.settings, settings),
        },
        hash(settings=$.defaults.transform.proc.hash.settings): {
          type: 'proc_hash',
          settings: std.mergePatch($.defaults.transform.proc.hash.settings, settings),
        },
        http(settings=$.defaults.transform.proc.http.settings): {
          type: 'proc_http',
          settings: std.mergePatch($.defaults.transform.proc.http.settings, settings),
        },
        insert(settings=$.defaults.transform.proc.insert.settings): {
          type: 'proc_insert',
          settings: std.mergePatch($.defaults.transform.proc.insert.settings, settings),
        },
        join(settings=$.defaults.transform.proc.join.settings): {
          type: 'proc_join',
          settings: std.mergePatch($.defaults.transform.proc.join.settings, settings),
        },
        jq(settings=$.defaults.transform.proc.jq.settings): {
          type: 'proc_jq',
          settings: std.mergePatch($.defaults.transform.proc.jq.settings, settings),
        },
        kv_store(settings=$.defaults.transform.proc.kv_store.settings): {
          type: 'proc_kv_store',
          settings: std.mergePatch($.defaults.transform.proc.kv_store.settings, settings),
        },
        math(settings=$.defaults.transform.proc.math.settings): {
          type: 'proc_math',
          settings: std.mergePatch($.defaults.transform.proc.math.settings, settings),
        },
        pretty_print(settings=$.defaults.transform.proc.pretty_print.settings): {
          type: 'proc_pretty_print',
          settings: std.mergePatch($.defaults.transform.proc.pretty_print.settings, settings),
        },
        replace(settings=$.defaults.transform.proc.replace.settings): {
          type: 'proc_replace',
          settings: std.mergePatch($.defaults.transform.proc.replace.settings, settings),
        },
        split(settings=$.defaults.transform.proc.split.settings): {
          type: 'proc_split',
          settings: std.mergePatch($.defaults.transform.proc.split.settings, settings),
        },
        time(settings=$.defaults.transform.proc.time.settings): {
          type: 'proc_time',
          settings: std.mergePatch($.defaults.transform.proc.time.settings, settings),
        },
      },
      send: {
        aws_dynamodb(settings=$.defaults.transform.send.aws_dynamodb.settings): {
          type: 'send_aws_dynamodb',
          settings: std.mergePatch($.defaults.transform.send.aws_dynamodb.settings, settings),
        },
        aws_kinesis(settings=$.defaults.transform.send.aws_kinesis.settings): {
          type: 'send_aws_kinesis',
          settings: std.mergePatch($.defaults.transform.send.aws_kinesis.settings, settings),
        },
        aws_kinesis_firehose(settings=$.defaults.transform.send.aws_kinesis_firehose.settings): {
          type: 'send_aws_kinesis_firehose',
          settings: std.mergePatch($.defaults.transform.send.aws_kinesis_firehose.settings, settings),
        },
        aws_s3(settings=$.defaults.transform.send.aws_s3.settings): {
          type: 'send_aws_s3',
          settings: std.mergePatch($.defaults.transform.send.aws_s3.settings, settings),
        },
        aws_sns(settings=$.defaults.transform.send.aws_sns.settings): {
          type: 'send_aws_sns',
          settings: std.mergePatch($.defaults.transform.send.aws_sns.settings, settings),
        },
        aws_sqs(settings=$.defaults.transform.send.aws_sqs.settings): {
          type: 'send_aws_sqs',
          settings: std.mergePatch($.defaults.transform.send.aws_sqs.settings, settings),
        },
        file(settings=$.defaults.transform.send.file.settings): {
          type: 'send_file',
          settings: std.mergePatch($.defaults.transform.send.file.settings, settings),
        },
        http(settings=$.defaults.transform.send.http.settings): {
          type: 'send_http',
          settings: std.mergePatch($.defaults.transform.send.http.settings, settings),
        },
        stdout: {
          type: 'send_stdout',
        },
        sumologic(settings=$.defaults.transform.send.sumologic.settings): {
          type: 'send_sumologic',
          settings: std.mergePatch($.defaults.transform.send.sumologic.settings, settings),
        },
      },
    },
    // Mirrors interfaces from the internal/kv_store package.
    kv_store: {
      aws_dynamodb(settings=$.defaults.kv_store.aws_dynamodb.settings): {
        local s = std.mergePatch($.defaults.kv_store.aws_dynamodb.settings, settings),
        type: 'aws_dynamodb',
        settings: std.mergePatch($.defaults.transform.proc_.settings, settings),
      },
      csv_file(settings=$.defaults.kv_store.csv_file.settings): {
        local s = std.mergePatch($.defaults.kv_store.csv_file.settings, settings),
        type: 'csv_file',
        settings: std.mergePatch($.defaults.transform.proc_.settings, settings),
      },
      json_file(settings=$.defaults.kv_store.json_file.settings): {
        local s = std.mergePatch($.defaults.kv_store.json_file.settings, settings),
        type: 'json_file',
        settings: std.mergePatch($.defaults.transform.proc_.settings, settings),
      },
      memory(settings=$.defaults.kv_store.memory.settings): {
        local s = std.mergePatch($.defaults.kv_store.memory.settings, settings),
        type: 'memory',
        settings: std.mergePatch($.defaults.transform.proc_.settings, settings),
      },
      mmdb(settings=$.defaults.kv_store.mmdb.settings): {
        local s = std.mergePatch($.defaults.kv_store.mmdb.settings, settings),
        type: 'mmdb',
        settings: std.mergePatch($.defaults.transform.proc_.settings, settings),
      },
      text_file(settings=$.defaults.kv_store.text_file.settings): {
        local s = std.mergePatch($.defaults.kv_store.text_file.settings, settings),
        type: 'text_file',
        settings: std.mergePatch($.defaults.transform.proc_.settings, settings),
      },
    },
  },
  patterns: {
    condition: {
      oper: {
        ip: {
          // Returns true if the key is a valid IP address and is not private.
          public(key=null): $.interfaces.condition.oper.none(
            $.patterns.condition.oper.ip.private(key=key).inspectors
            + [
              // The none operator combined with negation returns true if the key is a valid IP.
              $.interfaces.condition.insp.ip(settings={ key: key, negate: true, type: 'valid' }),
            ]
          ),
          // Returns true if the key is a private IP address.
          private(key=null): $.interfaces.condition.oper.any($.patterns.condition.insp.ip.private(key=key)),
        },
      },
      insp: {
        // Negates any inspector.
        negate(inspector): std.mergePatch(inspector, { settings: { negate: true } }),
        ip: {
          // Checks if an IP address is private.
          //
          // Use with the ANY operator to match private IP addresses.
          // Use with the NONE operator to match public IP addresses.
          private(key=null): [
            $.interfaces.condition.insp.ip(settings={ key: key, type: 'loopback' }),
            $.interfaces.condition.insp.ip(settings={ key: key, type: 'multicast' }),
            $.interfaces.condition.insp.ip(settings={ key: key, type: 'multicast_link_local' }),
            $.interfaces.condition.insp.ip(settings={ key: key, type: 'private' }),
            $.interfaces.condition.insp.ip(settings={ key: key, type: 'unicast_link_local' }),
            $.interfaces.condition.insp.ip(settings={ key: key, type: 'unspecified' }),
          ],
        },
        length: {
          // Checks if data is equal to zero.
          //
          // Use with the ANY / ALL operator to match empty data.
          // Use with the NONE operator to match non-empty data.
          eq_zero(key=null):
            $.interfaces.condition.insp.length(settings={ key: key, type: 'equals', length: 0 }),
          // checks if data is greater than zero.
          //
          // use with the ANY / ALL operator to match non-empty data.
          // use with the NONE operator to match empty data.
          gt_zero(key=null):
            $.interfaces.condition.insp.length(settings={ key: key, type: 'greater_than', length: 0 }),
        },
        string: {
          contains(string, key=null):
            $.interfaces.condition.insp.string(settings={ key: key, type: 'contains', string: string }),
          equals(string, key=null):
            $.interfaces.condition.insp.string(settings={ key: key, type: 'equals', string: string }),
          starts_with(string, key=null):
            $.interfaces.condition.insp.string(settings={ key: key, type: 'starts_with', string: string }),
          ends_with(string, key=null):
            $.interfaces.condition.insp.string(settings={ key: key, type: 'ends_with', string: string }),
        },
      },
    },
    transform: {
      // Conditional applies a transform when a single condition is met.
      conditional(transform, condition): {
        type: 'meta_switch',
        settings: { switch: [{ condition: condition, transform: transform }] },
      },
    },
  },
}
