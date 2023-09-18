{
  // Mirrors interfaces from the condition package.
  condition: {
    // Operators.
    all(i): { operator: 'all', inspectors: $.helpers.make_array(i) },
    any(i): { operator: 'any', inspectors: $.helpers.make_array(i) },
    none(i): { operator: 'none', inspectors: $.helpers.make_array(i) },
    // Inspectors.
    format: {
      content(settings=null): {
        local default = {
          object: $.config.object,
          type: null,
        },

        type: 'format_content',
        settings: std.mergePatch(default, settings),
      },
      json(settings=null): {
        type: 'format_json',
      },
    },
    logic: {
      len: {
        default: {
          object: $.config.object,
          length: null,
          measurement: 'byte',
        },
        equal_to(settings=null): {
          local default = $.condition.logic.len.default,

          type: 'logic_len_equal_to',
          settings: std.mergePatch(default, settings),
        },
        greater_than(settings=null): {
          local default = $.condition.logic.len.default,

          type: 'logic_len_greater_than',
          settings: std.mergePatch(default, settings),
        },
        less_than(settings=null): {
          local default = $.condition.logic.len.default,

          type: 'logic_len_less_than',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    meta: {
      condition(settings=null): {
        local default = { condition: null },

        type: 'meta_condition',
        settings: std.mergePatch(default, settings),
      },
      for_each(settings=null): {
        local default = {
          object: $.config.object,
          type: null,
          inspector: null,
        },

        type: 'meta_for_each',
        settings: std.mergePatch(default, settings),
      },
      negate(settings=null): {
        local default = { inspector: null },

        type: 'meta_negate',
        settings: std.mergePatch(default, settings),
      },
    },
    network: {
      ip: {
        default: {
          object: $.config.object,
        },
        global_unicast(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_global_unicast',
          settings: std.mergePatch(default, settings),
        },
        link_local_multicast(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_link_local_multicast',
          settings: std.mergePatch(default, settings),
        },
        link_local_unicast(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_link_local_unicast',
          settings: std.mergePatch(default, settings),
        },
        loopback(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_loopback',
          settings: std.mergePatch(default, settings),
        },
        multicast(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_multicast',
          settings: std.mergePatch(default, settings),
        },
        private(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_private',
          settings: std.mergePatch(default, settings),
        },
        unicast(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_unicast',
          settings: std.mergePatch(default, settings),
        },
        unspecified(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_unspecified',
          settings: std.mergePatch(default, settings),
        },
        valid(settings=null): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_valid',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    string: {
      default: {
        object: $.config.object,
        string: null,
      },
      contains(settings=null): {
        local default = $.condition.string.default,

        type: 'string_contains',
        settings: std.mergePatch(default, settings),
      },
      equal_to(settings=null): {
        local default = $.condition.string.default,

        type: 'string_equal_to',
        settings: std.mergePatch(default, settings),
      },
      starts_with(settings=null): {
        local default = $.condition.string.default,

        type: 'string_starts_with',
        settings: std.mergePatch(default, settings),
      },
      ends_with(settings=null): {
        local default = $.condition.string.default,

        type: 'string_ends_with',
        settings: std.mergePatch(default, settings),
      },
      pattern(settings=null): {
        local default = {
          object: $.config.object,
          pattern: null,
        },

        type: 'string_pattern',
        settings: std.mergePatch(default, settings),
      },
    },
    utility: {
      random(settings=null): {
        type: 'utility_random',
      },
    },
  },
  // Mirrors interfaces from the transform package.
  transform: {
    aggregate: {
      from: {
        array(settings=null): {
          local default = {
            object: $.config.object,
            buffer: $.config.buffer,
          },

          type: 'aggregate_from_array',
          settings: std.mergePatch(default, settings),
        },
        str(settings=null): {
          local default = {
            object: $.config.object,
            separator: null,
          },

          type: 'aggregate_from_str',
          settings: std.mergePatch(default, settings),
        },
      },
      to: {
        array(settings=null): {
          local default = {
            object: $.config.object,
            buffer: $.config.buffer,
          },

          type: 'aggregate_to_array',
          settings: std.mergePatch(default, settings),
        },
        str(settings=null): {
          local default = {
            object: $.config.object,
            separator: null,
          },

          type: 'aggregate_to_str',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    array: {
      group(settings=null): {
        local default = {
          object: $.config.object,
          group_keys: null,
        },

        type: 'array_group',
        settings: std.mergePatch(default, settings),
      },
      join(settings=null): {
        local default = {
          object: $.config.object,
          separator: null,
        },

        type: 'array_join',
        settings: std.mergePatch(default, settings),
      },
    },
    compress: {
      from_gzip(settings=null): {
        type: 'compress_from_gzip',
      },
      to_gzip(settings=null): {
        type: 'compress_to_gzip',
      },
    },
    enrich: {
      aws: {
        dynamodb(settings=null): {
          local default = {
            object: $.config.object,
            aws: $.config.aws,
            retry: $.config.retry,
            table: null,
            partition_key: null,
            sort_key: null,
            key_condition_expression: null,
            limit: 1,
            scan_index_forward: false,
          },

          type: 'enrich_aws_dynamodb',
          settings: std.mergePatch(default, settings),
        },
        lambda(settings=null): {
          local default = {
            object: $.config.object,
            aws: $.config.aws,
            retry: $.config.retry,
            function_name: null,
          },

          type: 'enrich_aws_lambda',
          settings: std.mergePatch(default, settings),
        },
      },
      dns: {
        default: {
          object: $.config.object,
          request: $.config.request,
        },
        domain_lookup(settings=null): {
          local default = $.transform.enrich.dns.default,

          type: 'enrich_dns_domain_lookup',
          settings: std.mergePatch(default, settings),
        },
        ip_lookup(settings=null): {
          local default = $.transform.enrich.dns.default,

          type: 'enrich_dns_ip_lookup',
          settings: std.mergePatch(default, settings),
        },
        txt_lookup(settings=null): {
          local default = $.transform.enrich.dns.default,

          type: 'enrich_dns_txt_lookup',
          settings: std.mergePatch(default, settings),
        },
      },
      http: {
        default: {
          object: $.config.object,
          request: $.config.request,
          url: null,
          headers: null,
        },
        get(settings=null): {
          local default = $.transform.http.default,

          type: 'enrich_http_get',
          settings: std.mergePatch(default, settings),
        },
        post(settings=null): {
          local default = $.transform.http.default { body_key: null },

          type: 'enrich_http_post',
          settings: std.mergePatch(default, settings),
        },
      },
      kv_store: {
        default: {
          object: $.config.object,
          prefix: null,
          kv_store: null,
          close_kv_store: false,
        },
        get(settings=null): {
          local default = $.transform.kv_store.default,

          type: 'enrich_kv_store_get',
          settings: std.mergePatch(default, settings),
        },
        set(settings=null): {
          local default = $.transform.kv_store.default { ttl_key: null, ttl_offset: null },

          type: 'enrich_kv_store_set',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    external: {
      jq(settings=null): {
        local default = { query: null },

        type: 'external_jq',
        settings: std.mergePatch(default, settings),
      },
    },
    format: {
      default: {
        object: $.config.object,
      },
      from_base64(settings=null): {
        local default = $.transform.format.default,

        type: 'format_from_base64',
        settings: std.mergePatch(default, settings),
      },
      from_pretty_print(settings=null): {
        type: 'format_from_pretty_print',
      },
      to_base64(settings=null): {
        local default = $.transform.format.default,

        type: 'format_to_base64',
        settings: std.mergePatch(default, settings),
      },
    },
    hash: {
      default: {
        object: $.config.object,
      },
      md5(settings=null): {
        local default = $.transform.hash.default,

        type: 'hash_md5',
        settings: std.mergePatch(default, settings),
      },
      sha256(settings=null): {
        local default = $.transform.hash.default,

        type: 'hash_sha256',
        settings: std.mergePatch(default, settings),
      },
    },
    logic: {
      num: {
        default: {
          object: $.config.object,
        },
        add(settings=null): {
          local default = $.transform.logic.num.default,

          type: 'logic_num_add',
          settings: std.mergePatch(default, settings),
        },
        subtract(settings=null): {
          local default = $.transform.logic.num.default,

          type: 'logic_num_subtract',
          settings: std.mergePatch(default, settings),
        },
        multiply(settings=null): {
          local default = $.transform.logic.num.default,

          type: 'logic_num_multiply',
          settings: std.mergePatch(default, settings),
        },
        divide(settings=null): {
          local default = $.transform.logic.num.default,

          type: 'logic_num_divide',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    meta: {
      for_each(settings=null): {
        local default = {
          object: $.config.object,
          transform: null,
        },

        type: 'meta_for_each',
        settings: std.mergePatch(default, settings),
      },
      pipeline(settings=null): {
        local default = {
          object: $.config.object,
          transform: null,
        },

        type: 'meta_pipeline',
        settings: std.mergePatch(default, settings),
      },
      switch(settings=null): {
        local default = { switch: null },

        type: 'meta_switch',
        settings: settings,
      },
    },
    network: {
      domain: {
        default: {
          object: $.config.object,
        },
        registered_domain(settings=null): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_registered_domain',
          settings: std.mergePatch(default, settings),
        },
        subdomain(settings=null): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_subdomain',
          settings: std.mergePatch(default, settings),
        },
        top_level_domain(settings=null): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_top_level_domain',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    object: {
      default: {
        object: $.config.object,
      },
      copy(settings=null): {
        local default = $.transform.object.default,

        type: 'object_copy',
        settings: std.mergePatch(default, settings),
      },
      delete(settings=null): {
        local default = $.transform.object.default,

        type: 'object_delete',
        settings: std.mergePatch(default, settings),
      },
      insert(settings=null): {
        local default = $.transform.object.default,

        type: 'object_insert',
        settings: std.mergePatch(default, settings),
      },
      to: {
        bool(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_bool',
          settings: std.mergePatch(default, settings),
        },
        float(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_float',
          settings: std.mergePatch(default, settings),
        },
        int(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_int',
          settings: std.mergePatch(default, settings),
        },
        str(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_str',
          settings: std.mergePatch(default, settings),
        },
        uint(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_uint',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    send: {
      aws: {
        dynamodb(settings=null): {
          local default = {
            aws: $.config.aws,
            retry: $.config.retry,
            table: null,
          },

          type: 'send_aws_dynamodb',
          settings: std.mergePatch(default, settings),
        },
        kinesis_data_firehose(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            stream: null,
          },

          type: 'send_aws_kinesis_data_firehose',
          settings: std.mergePatch(default, settings),
        },
        kinesis_data_stream(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            stream: null,
            partition: null,
            partition_key: null,
            shard_redistribution: false,
          },

          type: 'send_aws_kinesis_data_stream',
          settings: std.mergePatch(default, settings),
        },
        s3(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            bucket: null,
            file_path: $.file_path,
            file_format: { type: 'json' },
            file_compression: { type: 'gzip' },
          },

          type: 'send_aws_s3',
          settings: std.mergePatch(default, settings),
        },
        sns(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            arn: null,
          },

          type: 'send_aws_sns',
          settings: std.mergePatch(default, settings),
        },
        sqs(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            queue: null,
          },

          type: 'send_aws_sqs',
          settings: std.mergePatch(default, settings),
        },
      },
      file(settings=null): {
        local default = {
          buffer: $.config.buffer,
          file_path: $.file_path,
          file_format: { type: 'json' },
          file_compression: { type: 'gzip' },
        },

        type: 'send_file',
        settings: std.mergePatch(default, settings),
      },
      http(settings=null): {
        local default = {
          url: null,
          headers: null,
          headers_key: null,
        },

        type: 'send_http',
        settings: std.mergePatch(default, settings),
      },
      stdout(settings=null): {
        type: 'send_stdout',
      },
      sumologic(settings=null): {
        local default = {
          buffer: $.config.buffer,
          url: null,
          category: null,
          category_key: null,
        },

        type: 'send_sumologic',
        settings: std.mergePatch(default, settings),
      },
    },
    string: {
      pattern: {
        default: {
          object: $.config.object,
          pattern: null,
        },
        find_all(settings=null): {
          local default = $.transform.string.pattern.default,

          type: 'string_pattern_find_all',
          settings: std.mergePatch(default, settings),
        },
        find(settings=null): {
          local default = $.transform.string.pattern.default,

          type: 'string_pattern_find',
          settings: std.mergePatch(default, settings),
        },
        named_group(settings=null): {
          local default = $.transform.string.pattern.default,

          type: 'string_pattern_named_group',
          settings: std.mergePatch(default, settings),
        },
      },
      replace(settings=null): {
        local default = {
          object: $.config.object,
          old: null,
          new: null,
          count: -1,
        },

        type: 'string_replace',
        settings: std.mergePatch(default, settings),
      },
      split(settings=null): {
        local default = {
          object: $.config.object,
          separator: null,
        },

        type: 'string_split',
        settings: std.mergePatch(default, settings),
      },
      to: {
        lower(settings=null): {
          type: 'string_to_lower',
        },
        upper(settings=null): {
          type: 'string_to_upper',
        },
        snake(settings=null): {
          type: 'string_to_snake',
        },
      },
    },
    time: {
      from_str(settings=null): {
        local default = {
          object: $.config.object,
          format: null,
          location: null,
        },

        type: 'time_from_str',
        settings: std.mergePatch(default, settings),
      },
      from_unix(settings=null): {
        local default = {
          object: $.config.object,
        },

        type: 'time_from_unix',
        settings: std.mergePatch(default, settings),
      },
      now(settings=null): {
        local default = {
          object: $.config.object,
        },

        type: 'time_now',
        settings: std.mergePatch(default, settings),
      },
      to_str(settings=null): {
        local default = {
          object: $.config.object,
          format: null,
          location: null,
        },

        type: 'time_to_str',
        settings: std.mergePatch(default, settings),
      },
      to_unix(settings=null): {
        local default = {
          object: $.config.object,
        },

        type: 'time_to_unix',
        settings: std.mergePatch(default, settings),
      },
    },
    utility: {
      delay(settings=null): {
        local default = {
          duration: null,
        },

        type: 'utility_delay',
        settings: std.mergePatch(default, settings),
      },
      drop(settings=null): {
        type: 'utility_drop',
      },
      err(settings=null): {
        local default = {
          object: $.config.object,
          message: null,
        },

        type: 'utility_error',
        settings: std.mergePatch(default, settings),
      },
    },
  },
  // Mirrors interfaces from the internal/kv_store package.
  kv_store: {
    aws_dynamodb(settings=null): {
      local default = { table: null, attributes: { partition_key: null, sort_key: null, value: null, ttl: null }, consistent_read: false },

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
  // Mirrors structs from the internal/config package.
  config: {
    aws: { region: null, assume_role: null },
    buffer: { count: 1000, size: 100000, duration: '5m', key: null },
    object: { key: null, set_key: null },
    request: { timeout: '1s' },
    retry: { count: 3 },
  },
  // Mirrors config from the internal/file package.
  file_path: { prefix: null, prefix_key: null, time_format: '2006/01/02', uuid: true, extension: true },
  helpers: {
    // If the input is not an array, then this returns it as an array
    make_array(i): if !std.isArray(i) then [i] else i,
    key: {
      // If key is foo and arr is bar, then result is foo.bar
      // If key is foo and arr is [bar, baz], then result is foo.bar.baz
      append(key, arr): std.join('.', $.helpers.make_array(key) + $.helpers.make_array(arr)),
      // If key is foo, then result is foo.-1
      append_array(key): key + '.-1',
      // If key is foo and e is 0, then result is foo.0
      get_element(key, e=0): std.join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
    },
  },
  patterns: {
    condition: {
      obj(key): {
        object: { key: key },
      },
      // Negates any inspector.
      negate(inspector): $.condition.meta.negate(settings={ inspector: inspector }),
      network: {
        ip: {
          // Checks if an IP address is internal.
          //
          // Use with the ANY operator to match internal IP addresses.
          // Use with the NONE operator to match external IP addresses.
          internal(key=null): [
            $.condition.network.ip.link_local_multicast(settings=$.patterns.condition.obj(key)),
            $.condition.network.ip.link_local_unicast(settings=$.patterns.condition.obj(key)),
            $.condition.network.ip.loopback(settings=$.patterns.condition.obj(key)),
            $.condition.network.ip.multicast(settings=$.patterns.condition.obj(key)),
            $.condition.network.ip.private(settings=$.patterns.condition.obj(key)),
            $.condition.network.ip.unspecified(settings=$.patterns.condition.obj(key)),
          ],
        },
      },
      logic: {
        len: {
          // Checks if data is equal to zero.
          //
          // Use with the ANY / ALL operator to match empty data.
          // Use with the NONE operator to match non-empty data.
          eq_zero(key=null): 
            $.condition.logic.len.equal_to(settings=$.patterns.condition.obj(key) { length: 0 }),
          // Checks if data is greater than zero.
          //
          // Use with the ANY / ALL operator to match non-empty data.
          // Use with the NONE operator to match empty data.
          gt_zero(key=null):
            $.condition.logic.len.greater_than(settings=$.patterns.condition.obj(key) { length: 0 }),
        },
      },
      string: {
        contains(string, key=null):
          $.condition.string.contains(settings=$.patterns.condition.obj(key) { string: string }),
        equal_to(string, key=null):
          $.condition.string.equal_to(settings=$.patterns.condition.obj(key) { string: string }),
        starts_with(string, key=null):
          $.condition.string.starts_with(settings=$.patterns.condition.obj(key) { string: string }),
        ends_with(string, key=null):
          $.condition.string.ends_with(settings=$.patterns.condition.obj(key) { string: string }),
      },
    },
    transform: {
      // Conditional applies a transform when a single condition is met.
      conditional(condition, transform): {
        type: 'meta_switch',
        settings: { switch: [{ condition: condition, transform: transform }] },
      },
    },
  },
}
