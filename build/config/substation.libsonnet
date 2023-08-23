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
        settings: std.mergePatch(settings, default),
      },
      for_each(settings=null): {
        local default = { key: null, negate: null, type: null, inspector: null },

        type: 'meta_for_each',
        settings: std.mergePatch(settings, default),
      },
    },
    insp: {
      default: { key: null, negate: null },
      content(settings=null): {
        local default = $.condition.insp.default { type: null },

        type: 'insp_content',
        settings: std.mergePatch(settings, default),
      },
      ip(settings=null): {
        local default = $.condition.insp.default { type: null },

        type: 'insp_ip',
        settings: std.mergePatch(settings, default),
      },
      json_valid(settings=null): {
        local default = {},

        type: 'insp_json_valid',
        settings: std.mergePatch(settings, default),
      },
      length(settings=null): {
        local default = $.condition.insp.default { type: null, length: null, measurement: 'byte' },

        type: 'insp_length',
        settings: std.mergePatch(settings, default),
      },
      random(settings=null): {
        local default = {},

        type: 'insp_random',
        settings: std.mergePatch(settings, default),
      },
      regexp(settings=null): {
        local default = $.condition.insp.default { expression: null },

        type: 'insp_regexp',
        settings: std.mergePatch(settings, default),
      },
      string(settings=null): {
        local default = $.condition.insp.default { type: null, string: null },

        type: 'insp_string',
        settings: std.mergePatch(settings, default),
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
        settings: std.mergePatch(settings, default),
      },
      pipeline(settings=null): {
        assert settings.transforms != null : 'meta_pipeline must contain transforms',
        assert std.isArray(settings.transforms) : 'meta_pipeline must contain an array of transforms',

        local default = $.transform.default { transforms: null },

        type: 'meta_pipeline',
        settings: std.mergePatch(settings, default),
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
        settings: std.mergePatch(settings, default),
      },
      aws_lambda(settings=null): {
        local default =
          $.transform.default
          + $.transform.aws_auth
          + $.transform.request
          + { error_on_failure: false, function_name: null },

        type: 'proc_aws_lambda',
        settings: std.mergePatch(settings, default),
      },
      base64(settings=null): {
        local default =
          $.transform.default
          { direction: null },

        type: 'proc_base64',
        settings: std.mergePatch(settings, default),
      },
      capture(settings=null): {
        local default =
          $.transform.default
          { expression: null, type: 'find', count: -1 },

        type: 'proc_capture',
        settings: std.mergePatch(settings, default),
      },
      case(settings=null): {
        local default =
          $.transform.default
          { type: null },

        type: 'proc_case',
        settings: std.mergePatch(settings, default),
      },
      combine(settings=null): {
        local default =
          $.transform.default
          { combine_key: null, separator: null, max_count: 1000, max_size: 10000 },

        type: 'proc_combine',
        settings: std.mergePatch(settings, default),
      },
      convert(settings=null): {
        local default =
          $.transform.default
          { type: null },

        type: 'proc_convert',
        settings: std.mergePatch(settings, default),
      },
      copy(settings=null): {
        local default = $.transform.default,

        type: 'proc_copy',
        settings: std.mergePatch(settings, default),
      },
      delete(settings=null): {
        local default = { key: null },

        type: 'proc_delete',
        settings: std.mergePatch(settings, default),
      },
      dns(settings=null): {
        local default =
          $.transform.default
          + $.transform.request
          + { type: null },

        type: 'proc_dns',
        settings: std.mergePatch(settings, default),
      },
      domain(settings=null): {
        local default =
          $.transform.default
          { error_on_failure: false, type: null },

        type: 'proc_domain',
        settings: std.mergePatch(settings, default),
      },
      drop(settings=null): {
        local default = {},

        type: 'proc_drop',
        settings: std.mergePatch(settings, default),
      },
      err(settings=null): {
        local default = { err: null },

        type: 'proc_err',
        settings: std.mergePatch(settings, default),
      },
      expand(settings=null): {
        local default = $.transform.default,

        type: 'proc_expand',
        settings: std.mergePatch(settings, default),
      },
      flatten_array(settings=null): {
        local default =
          $.transform.default
          { deep: true },

        type: 'proc_flatten_array',
        settings: std.mergePatch(settings, default),
      },
      group(settings=null): {
        local default =
          $.transform.default
          { keys: null },

        type: 'proc_group',
        settings: std.mergePatch(settings, default),
      },
      gzip(settings=null): {
        local default =
          $.transform.default
          { direction: null },

        type: 'proc_gzip',
        settings: std.mergePatch(settings, default),
      },
      hash(settings=null): {
        local default =
          $.transform.default
          { algorithm: 'sha256' },

        type: 'proc_hash',
        settings: std.mergePatch(settings, default),
      },
      http(settings=null): {
        local default =
          $.transform.default
          { error_on_failure: false, method: 'get', url: null, headers: null, body_key: null },

        type: 'proc_http',
        settings: std.mergePatch(settings, default),
      },
      insert(settings=null): {
        local default = { set_key: null, value: null },

        type: 'proc_insert',
        settings: std.mergePatch(settings, default),
      },
      join(settings=null): {
        local default =
          $.transform.default
          { separator: null },

        type: 'proc_join',
        settings: std.mergePatch(settings, default),
      },
      jq(settings=null): {
        local default =
          $.transform.default
          { query: null },

        type: 'proc_jq',
        settings: std.mergePatch(settings, default),
      },
      kv_store(settings=null): {
        local default =
          $.transform.default
          { type: null, prefix: null, ttl_key: null, ttl_offset: null, kv_store: null },

        type: 'proc_kv_store',
        settings: std.mergePatch(settings, default),
      },
      math(settings=null): {
        local default =
          $.transform.default
          { operation: null },

        type: 'proc_math',
        settings: std.mergePatch(settings, default),
      },
      pretty_print(settings=null): {
        local default = { direction: null },

        type: 'proc_pretty_print',
        settings: std.mergePatch(settings, default),
      },
      replace(settings=null): {
        local default =
          $.transform.default
          { old: null, new: null, count: -1 },

        type: 'proc_replace',
        settings: std.mergePatch(settings, default),
      },
      split(settings=null): {
        local default =
          $.transform.default
          { separator: null },

        type: 'proc_split',
        settings: std.mergePatch(settings, default),
      },
      time(settings=null): {
        local default =
          $.transform.default
          { format: null, location: null, set_format: '2006-01-02T15:04:05.000000Z', set_location: null },

        type: 'proc_time',
        settings: std.mergePatch(settings, default),
      },
    },
    send: {
      aws_dynamodb(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { table: null, key: null },

        type: 'send_aws_dynamodb',
        settings: std.mergePatch(settings, default),
      },
      aws_kinesis(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { stream: null, partition: null, partition_key: null, shard_redistribution: false },

        type: 'send_aws_kinesis',
        settings: std.mergePatch(settings, default),
      },
      aws_kinesis_firehose(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { stream: null },

        type: 'send_aws_kinesis_firehose',
        settings: std.mergePatch(settings, default),
      },
      aws_s3(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { bucket: null, file_path: null, file_format: null, file_compression: null },

        type: 'send_aws_s3',
        settings: std.mergePatch(settings, default),
      },
      aws_sns(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { topic: null },

        type: 'send_aws_sns',
        settings: std.mergePatch(settings, default),
      },
      aws_sqs(settings=null): {
        local default =
          $.transform.aws_auth
          + $.transform.request
          + { queue: null },

        type: 'send_aws_sqs',
        settings: std.mergePatch(settings, default),
      },
      file(settings=null): {
        local default = { 
          buffer: $.aggregate,
          file_path: $.file_path,
          file_format: { type: 'json' }, 
          file_compression: { type: 'gzip' } 
        },

        type: 'send_file',
        settings: std.mergePatch(settings, default),
      },
      http(settings=null): {
        local default = { url: null, headers: null, headers_key: null },

        type: 'send_http',
        settings: std.mergePatch(settings, default),
      },
      stdout(settings=null): {
        local default = { },

        type: 'send_stdout',
        settings: std.mergePatch(settings, default),
      },
      sumologic(settings=null): {
        local default = { url: null, category: null, category_key: null },

        type: 'send_sumologic',
        settings: std.mergePatch(settings, default),
      },
    },
  },
  // Mirrors interfaces from the internal/kv_store package.
  kv_store: {
    aws_dynamodb(settings=null): {
      local default = { table: null, attributes: { partition_key: null, sort_key: null, value: null, ttl: null } },

      type: 'aws_dynamodb',
      settings: std.mergePatch(settings, default),
    },
    csv_file(settings=null): {
      local default = { file: null, column: null, delimiter: ',', header: null },

      type: 'csv_file',
      settings: std.mergePatch(settings, default),
    },
    json_file(settings=$.defaults.kv_store.json_file.settings): {
      local default = { file: null, is_lines: false },

      type: 'json_file',
      settings: std.mergePatch(settings, default),
    },
    memory(settings=null): {
      local default = { capacity: 1024 },

      type: 'memory',
      settings: std.mergePatch(settings, default),
    },
    mmdb(settings=null): {
      local default = { file: null },

      type: 'mmdb',
      settings: std.mergePatch(settings, default),
    },
    text_file(settings=null): {
      local default = { file: null },

      type: 'text_file',
      settings: std.mergePatch(settings, default),
    },
  },
  // Mirrors config from the internal/aggregate package.
  aggregate: {
    max_count: 0,
    max_size: 0,
    max_interval: 0,
  },
  // Mirrors config from the internal/file package.
  file_path: {
    prefix: null,
    prefix_key: null,
    time_format: '2006/01/02',
    extension: true,
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
