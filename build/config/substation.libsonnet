{
  // Mirrors interfaces from the condition package.
  cnd: $.condition,
  condition: {
    // Operators.
    all(i): { operator: 'all', inspectors: $.helpers.make_array(i) },
    any(i): { operator: 'any', inspectors: $.helpers.make_array(i) },
    none(i): { operator: 'none', inspectors: $.helpers.make_array(i) },
    // Inspectors.
    fmt: $.condition.format,
    format: {
      json(settings=null): {
        type: 'format_json',
      },
      mime(settings=null): {
        local default = {
          object: $.config.object,
          type: null,
        },

        type: 'format_mime',
        settings: std.mergePatch(default, settings),
      },
    },
    num: $.condition.number,
    number: {
      bitwise: {
        and(settings=null): {
          local default = {
            object: $.config.object,
            operand: null,
          },

          type: 'number_bitwise_and',
          settings: std.mergePatch(default, settings),
        },
        not(settings=null): {
          local default = {
            object: $.config.object,
          },

          type: 'number_bitwise_not',
          settings: std.mergePatch(default, settings),
        },
        or(settings=null): {
          local default = {
            object: $.config.object,
            operand: null,
          },

          type: 'number_bitwise_or',
          settings: std.mergePatch(default, settings),
        },
        xor(settings=null): {
          local default = {
            object: $.config.object,
            operand: null,
          },

          type: 'number_bitwise_xor',
          settings: std.mergePatch(default, settings),
        },
      },
      len: $.condition.number.length,
      length: {
        default: {
          object: $.config.object,
          length: null,
          measurement: 'byte',
        },
        eq(settings=null): $.condition.number.length.equal_to(settings=settings),
        equal_to(settings=null): {
          local default = $.condition.number.length.default,

          type: 'number_length_equal_to',
          settings: std.mergePatch(default, settings),
        },
        gt(settings=null): $.condition.number.length.greater_than(settings=settings),
        greater_than(settings=null): {
          local default = $.condition.number.length.default,

          type: 'number_length_greater_than',
          settings: std.mergePatch(default, settings),
        },
        lt(settings=null): $.condition.number.length.less_than(settings=settings),
        less_than(settings=null): {
          local default = $.condition.number.length.default,

          type: 'number_length_less_than',
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
    net: $.condition.network,
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
    str: $.condition.string,
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
      eq(settings=null): $.condition.string.equal_to(settings=settings),
      equal_to(settings=null): {
        local default = $.condition.string.default,

        type: 'string_equal_to',
        settings: std.mergePatch(default, settings),
      },
      gt(settings=null): $.condition.string.greater_than(settings=settings),
      greater_than(settings=null): {
        local default = $.condition.string.default,

        type: 'string_greater_than',
        settings: std.mergePatch(default, settings),
      },
      lt(settings=null): $.condition.string.less_than(settings=settings),
      less_than(settings=null): {
        local default = $.condition.string.default,

        type: 'string_less_than',
        settings: std.mergePatch(default, settings),
      },
      prefix(settings=null): $.condition.string.starts_with(settings=settings),
      starts_with(settings=null): {
        local default = $.condition.string.default,

        type: 'string_starts_with',
        settings: std.mergePatch(default, settings),
      },
      suffix(settings=null): $.condition.string.ends_with(settings=settings),
      ends_with(settings=null): {
        local default = $.condition.string.default,

        type: 'string_ends_with',
        settings: std.mergePatch(default, settings),
      },
      match(settings=null): {
        local default = {
          object: $.config.object,
          pattern: null,
        },

        type: 'string_match',
        settings: std.mergePatch(default, settings),
      },
    },
    util: $.transform.utility,
    utility: {
      random(settings=null): {
        type: 'utility_random',
      },
    },
  },
  // Mirrors interfaces from the transform package.
  tf: $.transform,
  transform: {
    agg: $.transform.aggregate,
    aggregate: {
      from: {
        array(settings=null): {
          local default = {
            object: $.config.object,
          },

          type: 'aggregate_from_array',
          settings: std.mergePatch(default, settings),
        },
        str(settings=null): $.transform.aggregate.from.string(settings=settings),
        string(settings=null): {
          local default = {
            separator: null,
          },

          type: 'aggregate_from_string',
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
        str(settings=null): $.transform.aggregate.to.string(settings=settings),
        string(settings=null): {
          local default = {
            buffer: $.config.buffer,
            separator: null,
          },

          type: 'aggregate_to_string',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    arr: $.transform.array,
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
    enrich: {
      aws: {
        dynamodb(settings=null): {
          local default = {
            object: $.config.object,
            aws: $.config.aws,
            retry: $.config.retry,
            table_name: null,
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
          local default = $.transform.enrich.http.default,

          type: 'enrich_http_get',
          settings: std.mergePatch(default, settings),
        },
        post(settings=null): {
          local default = $.transform.enrich.http.default { body_key: null },

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
          local default = $.transform.enrich.kv_store.default,

          type: 'enrich_kv_store_get',
          settings: std.mergePatch(default, settings),
        },
        set(settings=null): {
          local default = $.transform.enrich.kv_store.default { ttl_key: null, ttl_offset: null },

          type: 'enrich_kv_store_set',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    fmt: $.transform.format,
    format: {
      default: {
        object: $.config.object,
      },
      from: {
        b64(settings=null): $.transform.format.from.base64(settings=settings),
        base64(settings=null): {
          local default = $.transform.format.default,

          type: 'format_from_base64',
          settings: std.mergePatch(default, settings),
        },
        gz(settings=null): $.transform.format.from.gzip(settings=settings),
        gzip(settings=null): {
          type: 'format_from_gzip',
        },
        pretty_print(settings=null): {
          type: 'format_from_pretty_print',
        },
      },
      to: {
        b64(settings=null): $.transform.format.to.base64(settings=settings),
        base64(settings=null): {
          local default = $.transform.format.default,

          type: 'format_to_base64',
          settings: std.mergePatch(default, settings),
        },
        gz(settings=null): $.transform.format.to.gzip(settings=settings),
        gzip(settings=null): {
          type: 'format_to_gzip',
        },
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
    num: $.transform.number,
    number: {
      math: {
        default: {
          object: $.config.object,
        },
        add(settings=null): $.transform.number.math.addition(settings=settings),
        addition(settings=null): {
          local default = $.transform.number.math.default,

          type: 'number_math_addition',
          settings: std.mergePatch(default, settings),
        },
        sub(settings=null): $.transform.number.math.subtraction(settings=settings),
        subtraction(settings=null): {
          local default = $.transform.number.math.default,

          type: 'number_math_subtraction',
          settings: std.mergePatch(default, settings),
        },
        mul(settings=null): $.transform.number.math.multiplication(settings=settings),
        multiplication(settings=null): {
          local default = $.transform.number.math.default,

          type: 'number_math_multiplication',
          settings: std.mergePatch(default, settings),
        },
        div(settings=null): $.transform.number.math.division(settings=settings),
        division(settings=null): {
          local default = $.transform.number.math.default,

          type: 'number_math_division',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    meta: {
      err(settings=null): {
        local default = { transform: null },

        type: 'meta_err',
        settings: std.mergePatch(default, settings),
      },
      for_each(settings=null): {
        local default = {
          object: $.config.object,
          transform: null,
        },

        type: 'meta_for_each',
        settings: std.mergePatch(default, settings),
      },
      metrics: {
        duration(settings=null): {
          local default = {
            name: null,
            attributes: null,
            destination: null,
            transform: null,
          },

          type: 'meta_metrics_duration',
          settings: std.mergePatch(default, settings),
        },
      },
      pipe(settings=null): $.transform.meta.pipeline(settings=settings),
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
    net: $.transform.network,
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
        tld(settings=null): $.transform.network.domain.top_level_domain(settings=settings),
        top_level_domain(settings=null): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_top_level_domain',
          settings: std.mergePatch(default, settings),
        },
      },
    },
    obj: $.transform.object,
    object: {
      default: {
        object: $.config.object,
      },
      cp(settings=null): $.transform.object.copy(settings=settings),
      copy(settings=null): {
        local default = $.transform.object.default,

        type: 'object_copy',
        settings: std.mergePatch(default, settings),
      },
      del(settings=null): $.transform.object.delete(settings=settings),
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
      jq(settings=null): {
        local default = { query: null },

        type: 'object_jq',
        settings: std.mergePatch(default, settings),
      },
      to: {
        bool(settings=null): $.transform.object.to.boolean(settings=settings),
        boolean(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_boolean',
          settings: std.mergePatch(default, settings),
        },
        float(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_float',
          settings: std.mergePatch(default, settings),
        },
        int(settings=null): $.transform.object.to.integer(settings=settings),
        integer(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_integer',
          settings: std.mergePatch(default, settings),
        },
        str(settings=null): $.transform.object.to.string(settings=settings),
        string(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_string',
          settings: std.mergePatch(default, settings),
        },
        uint(settings=null): $.transform.object.to.unsigned_integer(settings=settings),
        unsigned_integer(settings=null): {
          local default = $.transform.object.default,

          type: 'object_to_unsigned_integer',
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
            table_name: null,
          },

          type: 'send_aws_dynamodb',
          settings: std.mergePatch(default, settings),
        },
        firehose(settings=null): $.transform.send.aws.kinesis_data_firehose(settings=settings),
        kinesis_data_firehose(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            stream_name: null,
          },

          type: 'send_aws_kinesis_data_firehose',
          settings: std.mergePatch(default, settings),
        },
        kinesis_data_stream(settings=null): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            stream_name: null,
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
            bucket_name: null,
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
            arn: null,
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
      http: {
        post(settings=null): {
          local default = {
            url: null,
            headers: null,
            headers_key: null,
          },

          type: 'send_http',
          settings: std.mergePatch(default, settings),
        },
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
    str: $.transform.string,
    string: {
      append(settings=null): {
        local default = {
          object: $.config.object,
          string: null,
        },

        type: 'string_append',
        settings: std.mergePatch(default, settings),
      },
      match: {
        default: {
          object: $.config.object,
          pattern: null,
        },
        find_all(settings=null): {
          local default = $.transform.string.pattern.default,

          type: 'string_match_find_all',
          settings: std.mergePatch(default, settings),
        },
        find(settings=null): {
          local default = $.transform.string.pattern.default,

          type: 'string_match_find',
          settings: std.mergePatch(default, settings),
        },
        named_group(settings=null): {
          local default = $.transform.string.pattern.default,

          type: 'string_match_named_group',
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
      uuid(settings=null): {
        local default = {
          object: $.config.object,
        },

        type: 'string_uuid',
        settings: std.mergePatch(default, settings),
      },
    },
    time: {
      from: {
        str(settings=null): $.transform.time.from.string(settings=settings),
        string(settings=null): {
          local default = {
            object: $.config.object,
            format: null,
            location: null,
          },

          type: 'time_from_string',
          settings: std.mergePatch(default, settings),
        },
        unix(settings=null): {
          local default = {
            object: $.config.object,
          },

          type: 'time_from_unix',
          settings: std.mergePatch(default, settings),
        },
      },
      now(settings=null): {
        local default = {
          object: $.config.object,
        },

        type: 'time_now',
        settings: std.mergePatch(default, settings),
      },
      to: {
        str(settings=null): $.transform.time.to.string(settings=settings),
        string(settings=null): {
          local default = {
            object: $.config.object,
            format: null,
            location: null,
          },

          type: 'time_to_string',
          settings: std.mergePatch(default, settings),
        },
      },
      unix(settings=null): {
        local default = {
          object: $.config.object,
        },

        type: 'time_to_unix',
        settings: std.mergePatch(default, settings),
      },
    },
    util: $.transform.utility,
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

        type: 'utility_err',
        settings: std.mergePatch(default, settings),
      },
      metrics: {
        count(settings=null): {
          local default = {
            name: null,
            attributes: null,
            destination: null,
          },

          type: 'utility_metrics_count',
          settings: std.mergePatch(default, settings),
        },
      },
      secret(settings=null): {
        local default = { secret: null },

        type: 'utility_secret',
        settings: std.mergePatch(default, settings),
      },
    },
  },
  // Mirrors interfaces from the internal/kv_store package.
  kv_store: {
    aws_dynamodb(settings=null): {
      local default = {
        aws: $.config.aws,
        retry: $.config.retry,
        table_name: null,
        attributes: { partition_key: null, sort_key: null, value: null, ttl: null },
        consistent_read: false,
      },

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
    aws: { region: null, assume_role_arn: null },
    buffer: { count: 1000, size: 100000, duration: '5m', key: null },
    object: { key: null, set_key: null },
    request: { timeout: '1s' },
    retry: { count: 3 },
  },
  // Mirrors config from the internal/file package.
  file_path: { prefix: null, prefix_key: null, time_format: '2006/01/02', uuid: true, extension: true },
  // Mirrors interfaces from the internal/secrets package.
  secrets: {
    default: { id: null, ttl: null },
    aws_secrets_manager(settings=null): {
      local default = {
        id: null,
        name: null,
        ttl_offset: 0,
        aws: $.config.aws,
        retry: $.config.retry,
      },

      type: 'aws_secrets_manager',
      settings: std.mergePatch(default, settings),
    },
    environment_variable(settings=null): {
      local default = { id: null, name: null, ttl_offset: 0 },

      type: 'environment_variable',
      settings: std.mergePatch(default, settings),
    },
  },
  // Commonly used condition and transform patterns.
  pattern: {
    cnd: $.pattern.condition,
    condition: {
      obj(key): {
        object: { key: key },
      },
      // Negates any inspector.
      negate(inspector): $.condition.meta.negate(settings={ inspector: inspector }),
      net: $.pattern.condition.network,
      network: {
        ip: {
          // Checks if an IP address is internal.
          //
          // Use with the ANY operator to match internal IP addresses.
          // Use with the NONE operator to match external IP addresses.
          internal(key=null): [
            $.condition.network.ip.link_local_multicast(settings=$.pattern.condition.obj(key)),
            $.condition.network.ip.link_local_unicast(settings=$.pattern.condition.obj(key)),
            $.condition.network.ip.loopback(settings=$.pattern.condition.obj(key)),
            $.condition.network.ip.multicast(settings=$.pattern.condition.obj(key)),
            $.condition.network.ip.private(settings=$.pattern.condition.obj(key)),
            $.condition.network.ip.unspecified(settings=$.pattern.condition.obj(key)),
          ],
        },
      },
      num: $.pattern.condition.number,
      number: {
        len: $.pattern.condition.number.length,
        length: {
          // Checks if data is equal to zero.
          //
          // Use with the ANY / ALL operator to match empty data.
          // Use with the NONE operator to match non-empty data.
          eq_zero(key=null):
            $.condition.number.length.equal_to(settings=$.pattern.condition.obj(key) { length: 0 }),
          // Checks if data is greater than zero.
          //
          // Use with the ANY / ALL operator to match non-empty data.
          // Use with the NONE operator to match empty data.
          gt_zero(key=null):
            $.condition.number.length.greater_than(settings=$.pattern.condition.obj(key) { length: 0 }),
        },
      },
    },
    tf: $.pattern.transform,
    transform: {
      // Conditional applies a transform when a single condition is met. If
      // the condition does not contain a valid operator, then it is assumed
      // to be an ANY operator.
      conditional(condition, transform): {
        local c = if std.objectHas(condition, 'type') then { operator: 'any', inspectors: [condition] } else condition,

        type: 'meta_switch',
        settings: { switch: [{ condition: c, transform: transform }] },
      },
    },
  },
  // Utility functions that can be used in conditions and transforms.
  helpers: {
    // If the input is not an array, then this returns it as an array.
    make_array(i): if !std.isArray(i) then [i] else i,
    key: {
      // If key is `foo` and arr is `bar`, then the result is `foo.bar`.
      // If key is `foo` and arr is `[bar, baz]`, then the result is `foo.bar.baz`.
      append(key, arr): std.join('.', $.helpers.make_array(key) + $.helpers.make_array(arr)),
      // If key is `foo`, then the result is `foo.-1`.
      append_array(key): key + '.-1',
      // If key is `foo` and e is `0`, then the result is `foo.0`.
      get_element(key, e=0): std.join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
    },
  },
}
