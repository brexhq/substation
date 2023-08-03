{
  helpers: {
    // if input is not an array, then this returns an array
    make_array(i): if !std.isArray(i) then [i] else i,
    key: {
      // if key is foo and arr is bar, then result is foo.bar
      // if key is foo and arr is [bar, baz], then result is foo.bar.baz
      append(key, arr): std.proc_join('.', $.helpers.make_array(key) + $.helpers.make_array(arr)),
      // if key is foo, then result is foo.-1
      append_array(key): key + '.-1',
      // if key is foo and e is 0, then result is foo.0
      get_element(key, e=0): std.proc_join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
    },
    inspector: {
      // validates base settings of any inspector by checking for the
      // existence of any fields except key and negate
      validate(settings): std.all([
        if !std.member(['key', 'negate'], x) then false else true
        for x in std.objectFields(settings)
      ]),
    },
  },
  defaults: {
    condition: {
      insp_content: {
        settings: { key: null, negate:null, type: null },
      },
      meta_condition: {
        settings: { negate: null, condition: null },
      },
      meta_for_each: {
        settings: { key: null, negate:null, type: null, inspector: null },
      },
      insp_ip: {
        settings: { key: null, negate:null, type: null },
      },
      insp_length: {
        settings: { key: null, negate:null, type: null, value: null, measurement: 'byte' },
      },
      insp_regexp: {
        settings: { key: null, negate:null, expression: null },
      },
      insp_strings: {
        settings: { key: null, negate:null, type: null, expression: null },
      },
    },
    transform: {
      meta_for_each: {
        settings: { key: null, set_key: null, transform: null },
      },
      meta_pipeline: {
        settings: { key: null, set_key: null, transforms: null },
      },
      meta_switch: {
        settings: { switch: null },
      },
      proc_aws_dynamodb: {
        settings: { key: null, set_key: null, table: null, key_condition_expression: null, limit: 1, scan_index_forward: false },
      },
      proc_aws_lambda: {
        settings: { key: null, set_key: null, function_name: null },
      },
      proc_base64: {
        settings: { key: null, set_key: null, direction: null },
      },
      proc_capture: {
        settings: { key: null, set_key: null, expression: null, type: 'find', count: -1 },
      },
      proc_case: {
        settings: { key: null, set_key: null, type: null },
      },
      proc_condense: {
        settings: { key: null, set_key: null, condense_key: null, separator: null, max_count: 1000, max_size: 10000 },
      },
      proc_convert: {
        settings: { key: null, set_key: null, type: null },
      },
      proc_copy: {
        settings: { key: null, set_key: null },
      },
      proc_dns: {
        settings: { key: null, set_key: null, type: null, timeout: 1000 },
      },
      proc_domain: {
        settings: { key: null, set_key: null, type: null },
      },
      proc_error: {
        settings: { err: null }
      },
      proc_flatten: {
        settings: { key: null, set_key: null, deep: true },
      },
      proc_group: {
        settings: { key: null, set_key: null, keys: null },
      },
      proc_gzip: {
        settings: { key: null, set_key: null, direction: null },
      },
      proc_hash: {
        settings: { key: null, set_key: null, algorithm: 'sha256' },
      },
      proc_http: {
        settings: { key: null, set_key: null, method: 'get', url: null, headers: null, body_key: null },
      },
      proc_insert: {
        settings: { set_key: null, value: null },
      },
      proc_join: {
        settings: { key: null, set_key: null, separator: null },
      },
      proc_jq: {
        settings: { query: null },
      },
      proc_kv_store: {
        settings: { key: null, set_key: null, type: null, prefix: null, ttl_key: null, offset_ttl: null, kv_options: null },
      },
      proc_math: {
        settings: { key: null, set_key: null, operation: null },
      },
      proc_pretty_print: {
        settings: { direction: null },
      },
      proc_replace: {
        settings: { key: null, set_key: null, old: null, new: null, count: -1 },
      },
      send_file: {
        settings: { file_path: null, file_format: { type: 'json' }, file_compression: { type: 'gzip' } },
      },
      send_stdout: {
        settings: null,
      },
      proc_split: {
        settings: { key: null, set_key: null, separator: null },
      },
      proc_time: {
        settings: { key: null, set_key: null, format: null, location: null, set_format: '2006-01-02T15:04:05.000000Z', set_location: null },
      },
      send_aws_dynamodb: {
        settings: { table: null, key: null },
      },
      send_aws_kinesis: {
        settings: { stream: null, partition: null, partition_key: null, shard_redistribution: false },
      },
      send_aws_kinesis_firehose: {
        settings: { stream: null },
      },
      send_aws_s3: {
        settings: { bucket: null, file_path: null, file_format: null, file_compression: null },
      },
      send_aws_sns: {
        settings: { arn: null },
      },
      send_aws_sqs: {
        settings: { queue: null },
      },
      send_http: {
        settings: { url: null, headers: null, headers_key: null },
      },
      send_sumologic: {
        settings: { url: null, category: null, category_key: null },
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
      oper_all(i): { operator: 'all', inspectors: if !std.isArray(i) then [i] else i },
      oper_any(i): { operator: 'any', inspectors: if !std.isArray(i) then [i] else i },
      oper_none(i): { operator: 'none', inspectors: if !std.isArray(i) then [i] else i },
      meta_condition(settings=$.defaults.condition.meta_condition.settings): {
        type: 'meta_condition',
        settings: std.mergePatch($.defaults.condition.meta_condition.settings, settings),
      },
      meta_for_each(settings=$.defaults.condition.meta_for_each.settings): {
        type: 'meta_for_each',
        settings: std.mergePatch($.defaults.condition.meta_for_each.settings, settings),
      },
      insp_content(settings=$.defaults.condition.insp_content.settings): {
        type: 'insp_content',
        settings: std.mergePatch($.defaults.condition.insp_content.settings, settings),
      },
      insp_ip(settings=$.defaults.condition.insp_ip.settings): {
        type: 'insp_ip',
        settings: std.mergePatch($.defaults.condition.insp_ip.settings, settings),
      },
      insp_json_valid(settings=$.defaults.condition.insp_json_valid.settings): {
        type: 'insp_json_valid',
        settings: std.mergePatch($.defaults.condition.insp_json_valid.settings, settings),
      },
      insp_length(settings=$.defaults.condition.insp_length.settings): {
        type: 'insp_length',
        settings: std.mergePatch($.defaults.condition.insp_length.settings, settings),
      },
      insp_random: {
        type: 'random',
      },
      insp_regexp(settings=$.defaults.condition.insp_regexp.settings): {
        type: 'insp_regexp',
        settings: std.mergePatch($.defaults.condition.insp_regexp.settings, settings),
      },
      insp_strings(settings=$.defaults.condition.insp_strings.settings): {
        type: 'insp_strings',
        settings: std.mergePatch($.defaults.condition.insp_strings.settings, settings),
      },
    },
    // Mirrors interfaces from the transform package.
    transform: {
      meta_for_each(settings=$.defaults.transform.meta_for_each.settings): {
        type: 'meta_for_each',
        settings: std.mergePatch($.defaults.transform.meta_for_each.settings, settings),
      },
      meta_pipeline(settings=$.defaults.transform.meta_pipeline.settings): {
        type: 'meta_pipeline',
        settings: std.mergePatch($.defaults.transform.meta_pipeline.settings, settings),
      },
      meta_switch(settings=$.defaults.transform.meta_switch.settings): {
        type: 'meta_switch',
        settings: std.mergePatch($.defaults.transform.meta_switch.settings, settings),
      },
      proc_condense(settings=$.defaults.transform.proc_condense.settings): {
        type: 'proc_condense',
        settings: std.mergePatch($.defaults.transform.proc_condense.settings, settings),
      },
      proc_aws_dynamodb(settings=$.defaults.transform.proc_aws_dynamodb.settings): {
        type: 'proc_aws_dynamodb',
        settings: std.mergePatch($.defaults.transform.proc_aws_dynamodb.settings, settings),
      },
      proc_aws_lambda(settings=$.defaults.transform.proc_aws_lambda.settings): {
        type: 'proc_aws_lambda',
        settings: std.mergePatch($.defaults.transform.proc_aws_lambda.settings, settings),
      },
      proc_base64(settings=$.defaults.transform.proc_base64.settings): {
        type: 'proc_base64',
        settings: std.mergePatch($.defaults.transform.proc_base64.settings, settings),
      },
      proc_capture(settings=$.defaults.transform.proc_capture.settings): {
        type: 'proc_capture',
        settings: std.mergePatch($.defaults.transform.proc_capture.settings, settings),
      },
      proc_case(settings=$.defaults.transform.proc_case.settings): {
        type: 'proc_case',
        settings: std.mergePatch($.defaults.transform.proc_case.settings, settings),
      },
      proc_convert(settings=$.defaults.transform.proc_convert.settings): {
        type: 'proc_convert',
        settings: std.mergePatch($.defaults.transform.proc_convert.settings, settings),
      },
      proc_copy(settings=$.defaults.transform.proc_copy.settings): {
        type: 'proc_copy',
        settings: std.mergePatch($.defaults.transform.proc_copy.settings, settings),
      },
      proc_delete(settings=$.defaults.transform.proc_delete.settings): {
        type: 'proc_delete',
        settings: std.mergePatch($.defaults.transform.proc_delete.settings, settings),
      },
      proc_dns(settings=$.defaults.transform.proc_dns.settings): {
        type: 'proc_dns',
        settings: std.mergePatch($.defaults.transform.proc_dns.settings, settings),
      },
      proc_domain(settings=$.defaults.transform.proc_domain.settings): {
        type: 'proc_domain',
        settings: std.mergePatch($.defaults.transform.proc_domain.settings, settings),
      },
      proc_drop(settings=$.defaults.transform.proc_drop.settings): {
        type: 'proc_drop',
        settings: std.mergePatch($.defaults.transform.proc_drop.settings, settings),
      },
      proc_expand(settings=$.defaults.transform.proc_expand.settings): {
        type: 'proc_expand',
        settings: std.mergePatch($.defaults.transform.proc_expand.settings, settings),
      },
      proc_flatten(settings=$.defaults.transform.proc_flatten.settings): {
        type: 'proc_flatten',
        settings: std.mergePatch($.defaults.transform.proc_flatten.settings, settings),
      },
      proc_group(settings=$.defaults.transform.proc_group.settings): {
        type: 'proc_group',
        settings: std.mergePatch($.defaults.transform.proc_group.settings, settings),
      },
      proc_gzip(settings=$.defaults.transform.proc_gzip.settings): {
        type: 'proc_gzip',
        settings: std.mergePatch($.defaults.transform.proc_gzip.settings, settings),
      },
      proc_hash(settings=$.defaults.transform.proc_hash.settings): {
        type: 'proc_hash',
        settings: std.mergePatch($.defaults.transform.proc_hash.settings, settings),
      },
      proc_http(settings=$.defaults.transform.proc_http.settings): {
        type: 'proc_http',
        settings: std.mergePatch($.defaults.transform.proc_http.settings, settings),
      },
      proc_insert(settings=$.defaults.transform.proc_insert.settings): {
        type: 'proc_insert',
        settings: std.mergePatch($.defaults.transform.proc_insert.settings, settings),
      },
      proc_join(settings=$.defaults.transform.proc_join.settings): {
        type: 'proc_join',
        settings: std.mergePatch($.defaults.transform.proc_join.settings, settings),
      },
      proc_jq(settings=$.defaults.transform.proc_jq.settings): {
        type: 'proc_jq',
        settings: std.mergePatch($.defaults.transform.proc_jq.settings, settings),
      },
      proc_kv_store(settings=$.defaults.transform.proc_kv_store.settings): {
        type: 'proc_kv_store',
        settings: std.mergePatch($.defaults.transform.proc_kv_store.settings, settings),
      },
      proc_math(settings=$.defaults.transform.proc_math.settings): {
        type: 'proc_math',
        settings: std.mergePatch($.defaults.transform.proc_math.settings, settings),
      },
      proc_pretty_print(settings=$.defaults.transform.proc_pretty_print.settings): {
        type: 'proc_pretty_print',
        settings: std.mergePatch($.defaults.transform.proc_pretty_print.settings, settings),
      },
      proc_replace(settings=$.defaults.transform.proc_replace.settings): {
        type: 'proc_replace',
        settings: std.mergePatch($.defaults.transform.proc_replace.settings, settings),
      },
      proc_split(settings=$.defaults.transform.proc_split.settings): {
        type: 'proc_split',
        settings: std.mergePatch($.defaults.transform.proc_split.settings, settings),
      },
      proc_time(settings=$.defaults.transform.proc_time.settings): {
        type: 'proc_time',
        settings: std.mergePatch($.defaults.transform.proc_time.settings, settings),
      },
      send_aws_dynamodb(settings=$.defaults.transform.send_aws_dynamodb.settings): {
        type: 'send_aws_dynamodb',
        settings: std.mergePatch($.defaults.transform.send_aws_dynamodb.settings, settings),
      },
      send_aws_kinesis(settings=$.defaults.transform.send_aws_kinesis.settings): {
        type: 'send_aws_kinesis',
        settings: std.mergePatch($.defaults.transform.send_aws_kinesis.settings, settings),
      },
      send_aws_kinesis_firehose(settings=$.defaults.transform.send_aws_kinesis_firehose.settings): {
        type: 'send_aws_kinesis_firehose',
        settings: std.mergePatch($.defaults.transform.send_aws_kinesis_firehose.settings, settings),
      },
      send_aws_s3(settings=$.defaults.transform.send_aws_s3.settings): {
        type: 'send_aws_s3',
        settings: std.mergePatch($.defaults.transform.send_aws_s3.settings, settings),
      },
      send_aws_sns(settings=$.defaults.transform.send_aws_sns.settings): {
        type: 'send_aws_sns',
        settings: std.mergePatch($.defaults.transform.send_aws_sns.settings, settings),
      },
      send_aws_sqs(settings=$.defaults.transform.send_aws_sqs.settings): {
        type: 'send_aws_sqs',
        settings: std.mergePatch($.defaults.transform.send_aws_sqs.settings, settings),
      },
      send_file(settings=$.defaults.transform.send_file.settings): {
        type: 'send_file',
        settings: std.mergePatch($.defaults.transform.send_file.settings, settings),
      },
      send_http(settings=$.defaults.transform.send_http.settings): {
        type: 'send_http',
        settings: std.mergePatch($.defaults.transform.send_http.settings, settings),
      },
      send_stdout: {
        type: 'send_stdout',
      },
      send_sumologic(settings=$.defaults.transform.send_sumologic.settings): {
        type: 'send_sumologic',
        settings: std.mergePatch($.defaults.transform.send_sumologic.settings, settings),
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
          public(key=null): $.interfaces.condition.oper_none(
            $.patterns.inspector.ip.private(key=key)
            + [
              // The none operator combined with negation returns true if the key is a valid IP.
              $.interfaces.condition.insp_ip(settings={ key: key, negate: true,  type: 'valid' }),
            ]
          ),
          // Returns true if the key is a private IP address.
          private(key=null): $.interfaces.condition.oper_any($.patterns.inspector.ip.private(key=key)),
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
            $.interfaces.condition.insp_ip(settings={ key: key, type: 'loopback' }),
            $.interfaces.condition.insp_ip(settings={ key: key, type: 'multicast' }),
            $.interfaces.condition.insp_ip(settings={ key: key, type: 'multicast_link_local' }),
            $.interfaces.condition.insp_ip(settings={ key: key, type: 'private' }),
            $.interfaces.condition.insp_ip(settings={ key: key, type: 'unicast_link_local' }),
            $.interfaces.condition.insp_ip(settings={ key: key, type: 'unspecified' }),
          ],
        },
        length: {
          // Checks if data is equal to zero.
          //
          // Use with the ANY / ALL operator to match empty data.
          // Use with the NONE operator to match non-empty data.
          eq_zero(key=null):
            $.interfaces.condition.insp_length(settings={ key: key, type: 'equals', value: 0 }),
          // checks if data is greater than zero.
          //
          // use with the ANY / ALL operator to match non-empty data.
          // use with the NONE operator to match empty data.
          gt_zero(key=null):
            $.interfaces.condition.insp_length(settings={ key: key, type: 'greater_than', value: 0 }),
        },
        strings: {
          contains(expression, key=null):
            $.interfaces.condition.insp_strings(settings={ key: key, type: 'contains', expression: expression }),
          equals(expression, key=null):
            $.interfaces.condition.insp_strings(settings={ key: key, type: 'equals', expression: expression }),
          starts_with(expression, key=null):
            $.interfaces.condition.insp_strings(settings={ key: key, type: 'starts_with', expression: expression }),
          ends_with(expression, key=null):
            $.interfaces.condition.insp_strings(settings={ key: key, type: 'ends_with', expression: expression }),
        }, 
      }
    },
    transform: {
      // Conditional applies a transform when a single condition is met.
      conditional(transform, condition): {
        type: 'meta_switch',
        settings: { switch: [{condition: condition, transform: transform}] },
      },
    },
  },
}
