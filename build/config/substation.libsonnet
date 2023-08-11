{
  // Mirrors interfaces from the condition package.
  condition: {
    oper: {
      all(i): { operator: 'all', inspectors: $.helpers.make_array(i) },
      any(i): { operator: 'any', inspectors: $.helpers.make_array(i) },
      none(i): { operator: 'none', inspectors: $.helpers.make_array(i) },
    },
    meta: {
      condition(settings=null): {
        local default = { negate: null, condition: null },

        type: 'meta_condition',
        settings: std.mergePatch(default, settings),
      },
      for_each(settings=null): {
        local default = { key: null, negate: null, type: null, inspector: null },

        type: 'meta_for_each',
        settings: std.mergePatch(default, settings),
      },
    },
    insp: {
      default: { key: null, negate: null },
      content(settings=null): {
        local default = $.condition.insp.default { type: null },

        type: 'insp_content',
        settings: std.mergePatch(default, settings),
      },
      ip(settings=null): {
        local default = $.condition.insp.default { type: null },

        type: 'insp_ip',
        settings: std.mergePatch(default, settings),
      },
      json_valid(settings=null): {
        local default = {},

        type: 'insp_json_valid',
        settings: std.mergePatch(default, settings),
      },
      length(settings=null): {
        local default = $.condition.insp.default { type: null, length: null, measurement: 'byte' },

        type: 'insp_length',
        settings: std.mergePatch(default, settings),
      },
      random(settings=null): {
        local default = {},

        type: 'insp_random',
        settings: std.mergePatch(default, settings),
      },
      regexp(settings=null): {
        local default = $.condition.insp.default { expression: null },

        type: 'insp_regexp',
        settings: std.mergePatch(default, settings),
      },
      string(settings=null): {
        local default = $.condition.insp.default { type: null, string: null },

        type: 'insp_string',
        settings: std.mergePatch(default, settings),
      },
    },
  },
  // Mirrors interfaces from the transform package.
  transform: {
    default: { key: null, set_key: null },
    aws_auth: { auth: { region: null, assume_role: null } },
    request: { request: { timeout: null, max_retries: 1 } },
    meta: {
      for_each(settings=null): {
        assert settings.transform.type != null && settings.transform.settings != null : 'meta_for_each must contain a transform',

        local default = $.transform.default { transform: null },

        type: 'meta_for_each',
        settings: std.mergePatch(default, settings),
      },
      pipeline(settings=null): {
        assert settings.transforms != null : 'meta_pipeline must contain transforms',
        assert std.isArray(settings.transforms) : 'meta_pipeline must contain an array of transforms',

        local default = $.transform.default { transforms: null },

        type: 'meta_pipeline',
        settings: std.mergePatch(default, settings),
      },
      switch(settings=null): {
        assert settings.switch != null : 'meta_switch must be set',
        assert std.isArray(settings.switch) : 'meta_switch must be an array',

        type: 'meta_switch',
        settings: settings,
      },
    },
    proc: {
      aws_dynamodb(settings=null): {
        local default =
          $.transform.default
          + $.transform.aws_auth
          + $.transform.request
          + { table: null, key: null, set_key: null, key_condition_expression: null, limit: 1, scan_index_forward: false },

        type: 'proc_aws_dynamodb',
        settings: std.mergePatch(default, settings),
      },
      aws_lambda(settings=null): {
        local default =
          $.transform.default
          + $.transform.aws_auth
          + $.transform.request
          + { error_on_failure: false, function_name: null },

        type: 'proc_aws_lambda',
        settings: std.mergePatch(default, settings),
      },
      base64(settings=null): {
        local default =
          $.transform.default
          { direction: null },

        type: 'proc_base64',
        settings: std.mergePatch(default, settings),
      },
      capture(settings=null): {
        local default =
          $.transform.default
          { expression: null, type: 'find', count: -1 },

        type: 'proc_capture',
        settings: std.mergePatch(default, settings),
      },
      case(settings=null): {
        local default =
          $.transform.default
          { type: null },

        type: 'proc_case',
        settings: std.mergePatch(default, settings),
      },
      combine(settings=null): {
        local default =
          $.transform.default
          { combine_key: null, separator: null, max_count: 1000, max_size: 10000 },

        type: 'proc_combine',
        settings: std.mergePatch(default, settings),
      },
      convert(settings=null): {
        local default =
          $.transform.default
          { type: null },

        type: 'proc_convert',
        settings: std.mergePatch(default, settings),
      },
      copy(settings=null): {
        local default = $.transform.default,

        type: 'proc_copy',
        settings: std.mergePatch(default, settings),
      },
      delete(settings=null): {
        local default = { key: null },

        type: 'proc_delete',
        settings: std.mergePatch(default, settings),
      },
      dns(settings=null): {
        local default =
          $.transform.default
          + $.transform.request
          + { type: null },

        type: 'proc_dns',
        settings: std.mergePatch(default, settings),
      },
      domain(settings=null): {
        local default =
          $.transform.default
          { error_on_failure: false, type: null },

        type: 'proc_domain',
        settings: std.mergePatch(default, settings),
      },
      drop(settings=null): {
        local default = {},

        type: 'proc_drop',
        settings: std.mergePatch(default, settings),
      },
      err(settings=null): {
        local default = { err: null },

        type: 'proc_err',
        settings: std.mergePatch(default, settings),
      },
      expand(settings=null): {
        local default = $.transform.default,

        type: 'proc_expand',
        settings: std.mergePatch(default, settings),
      },
      flatten_array(settings=null): {
        local default =
          $.transform.default
          { deep: true },

        type: 'proc_flatten_array',
        settings: std.mergePatch(default, settings),
      },
      group(settings=null): {
        local default =
          $.transform.default
          { keys: null },

        type: 'proc_group',
        settings: std.mergePatch(default, settings),
      },
      gzip(settings=null): {
        local default =
          $.transform.default
          { direction: null },

        type: 'proc_gzip',
        settings: std.mergePatch(default, settings),
      },
      hash(settings=null): {
        local default =
          $.transform.default
          { algorithm: 'sha256' },

        type: 'proc_hash',
        settings: std.mergePatch(default, settings),
      },
      http(settings=null): {
        local default =
          $.transform.default
          { error_on_failure: false, method: 'get', url: null, headers: null, body_key: null },

        type: 'proc_http',
        settings: std.mergePatch(default, settings),
      },
      insert(settings=null): {
        local default = { set_key: null, value: null },

        type: 'proc_insert',
        settings: std.mergePatch(default, settings),
      },
      join(settings=null): {
        local default =
          $.transform.default
          { separator: null },

        type: 'proc_join',
        settings: std.mergePatch(default, settings),
      },
      jq(settings=null): {
        local default =
          $.transform.default
          { query: null },

        type: 'proc_jq',
        settings: std.mergePatch(default, settings),
      },
      kv_store(settings=null): {
        local default =
          $.transform.default
          { type: null, prefix: null, ttl_key: null, ttl_offset: null, kv_store: null },

        type: 'proc_kv_store',
        settings: std.mergePatch(default, settings),
      },
      math(settings=null): {
        local default =
          $.transform.default
          { operation: null },

        type: 'proc_math',
        settings: std.mergePatch(default, settings),
      },
      pretty_print(settings=null): {
        local default = { direction: null },

        type: 'proc_pretty_print',
        settings: std.mergePatch(default, settings),
      },
      replace(settings=null): {
        local default =
          $.transform.default
          { old: null, new: null, count: -1 },

        type: 'proc_replace',
        settings: std.mergePatch(default, settings),
      },
      split(settings=null): {
        local default =
          $.transform.default
          { separator: null },

        type: 'proc_split',
        settings: std.mergePatch(default, settings),
      },
      time(settings=null): {
        local default =
          $.transform.default
          { format: null, location: null, set_format: '2006-01-02T15:04:05.000000Z', set_location: null },

        type: 'proc_time',
        settings: std.mergePatch(default, settings),
      },
    },
    send: {
      aws_dynamodb(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { table: null, key: null },

        type: 'send_aws_dynamodb',
        settings: std.mergePatch(default, settings),
      },
      aws_kinesis(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { stream: null, partition: null, partition_key: null, shard_redistribution: false },

        type: 'send_aws_kinesis',
        settings: std.mergePatch(default, settings),
      },
      aws_kinesis_firehose(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { stream: null },

        type: 'send_aws_kinesis_firehose',
        settings: std.mergePatch(default, settings),
      },
      aws_s3(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { bucket: null, file_path: null, file_format: null, file_compression: null },

        type: 'send_aws_s3',
        settings: std.mergePatch(default, settings),
      },
      aws_sns(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { topic: null },

        type: 'send_aws_sns',
        settings: std.mergePatch(default, settings),
      },
      aws_sqs(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { queue: null },

        type: 'send_aws_sqs',
        settings: std.mergePatch(default, settings),
      },
      file(settings=null): {
        local default = { file_path: null, file_format: { type: 'json' }, file_compression: { type: 'gzip' } },

        type: 'send_file',
        settings: std.mergePatch(default, settings),
      },
      http(settings=null): {
        local default = { url: null, headers: null, headers_key: null },

        type: 'send_http',
        settings: std.mergePatch(default, settings),
      },
      stdout: {
        type: 'send_stdout',
      },
      sumologic(settings=null): {
        local default = { url: null, category: null, category_key: null },

        type: 'send_sumologic',
        settings: std.mergePatch(default, settings),
      },
    },
  },
  // Mirrors interfaces from the internal/kv_store package.
  kv_store: {
    aws_dynamodb(settings=null): {
      local default = { table: null, attributes: { partition_key: null, sort_key: null, value: null, ttl: null } },

      type: 'aws_dynamodb',
      settings: std.mergePatch(default, settings),
    },
    csv_file(settings=null): {
      local default = { file: null, column: null, delimiter: ',', header: null },

      type: 'csv_file',
      settings: std.mergePatch(default, settings),
    },
    json_file(settings=$.defaults.kv_store.json_file.settings): {
      local default = { file: null, is_lines: false },

      type: 'json_file',
      settings: std.mergePatch(default, settings),
    },
    memory(settings=null): {
      local default = { capacity: 1024 },

      type: 'memory',
      settings: std.mergePatch(default, settings),
    },
    mmdb(settings=null): {
      local default = { file: null },

      type: 'mmdb',
      settings: std.mergePatch(default, settings),
    },
    text_file(settings=null): {
      local default = { file: null },

      type: 'text_file',
      settings: std.mergePatch(default, settings),
    },
  },
  helpers: {
    // If the input is not an array, then this returns it as an array.
    make_array(i): if !std.isArray(i) then [i] else i,
    key: {
      // If key is foo and arr is bar, then result is foo.bar.
      // If key is foo and arr is [bar, baz], then result is foo.bar.baz.
      append(key, arr): std.join('.', $.helpers.make_array(key) + $.helpers.make_array(arr)),
      // if key is foo, then result is foo.-1
      append_array(key): key + '.-1',
      // if key is foo and e is 0, then result is foo.0
      get_element(key, e=0): std.join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
    },
  },
  patterns: {
    condition: {
      oper: {
        ip: {
          // Returns true if the key is a valid IP address and is not private.
          public(key=null): $.condition.oper.none(
            $.patterns.condition.oper.ip.private(key=key).inspectors
            + [
              // The none operator combined with negation returns true if the key is a valid IP.
              $.condition.insp.ip(settings={ key: key, negate: true, type: 'valid' }),
            ]
          ),
          // Returns true if the key is a private IP address.
          private(key=null): $.condition.oper.any($.patterns.condition.insp.ip.private(key=key)),
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
            $.condition.insp.ip(settings={ key: key, type: 'loopback' }),
            $.condition.insp.ip(settings={ key: key, type: 'multicast' }),
            $.condition.insp.ip(settings={ key: key, type: 'multicast_link_local' }),
            $.condition.insp.ip(settings={ key: key, type: 'private' }),
            $.condition.insp.ip(settings={ key: key, type: 'unicast_link_local' }),
            $.condition.insp.ip(settings={ key: key, type: 'unspecified' }),
          ],
        },
        length: {
          // Checks if data is equal to zero.
          //
          // Use with the ANY / ALL operator to match empty data.
          // Use with the NONE operator to match non-empty data.
          eq_zero(key=null):
            $.condition.insp.length(settings={ key: key, type: 'equals', length: 0 }),
          // checks if data is greater than zero.
          //
          // use with the ANY / ALL operator to match non-empty data.
          // use with the NONE operator to match empty data.
          gt_zero(key=null):
            $.condition.insp.length(settings={ key: key, type: 'greater_than', length: 0 }),
        },
        string: {
          contains(string, key=null):
            $.condition.insp.string(settings={ key: key, type: 'contains', string: string }),
          equals(string, key=null):
            $.condition.insp.string(settings={ key: key, type: 'equals', string: string }),
          starts_with(string, key=null):
            $.condition.insp.string(settings={ key: key, type: 'starts_with', string: string }),
          ends_with(string, key=null):
            $.condition.insp.string(settings={ key: key, type: 'ends_with', string: string }),
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
