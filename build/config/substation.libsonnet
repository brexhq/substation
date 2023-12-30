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
      json(settings={}): {
        type: 'format_json',
      },
      mime(settings={}): {
        local default = {
          object: $.config.object,
          type: null,
        },

        type: 'format_mime',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    num: $.condition.number,
    number: {
      bitwise: {
        and(settings={}): {
          local default = {
            object: $.config.object,
            value: null,
          },

          type: 'number_bitwise_and',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        not(settings={}): {
          local default = {
            object: $.config.object,
          },

          type: 'number_bitwise_not',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        or(settings={}): {
          local default = {
            object: $.config.object,
            value: null,
          },

          type: 'number_bitwise_or',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        xor(settings={}): {
          local default = {
            object: $.config.object,
            value: null,
          },

          type: 'number_bitwise_xor',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      len: $.condition.number.length,
      length: {
        default: {
          object: $.config.object,
          value: null,
          measurement: 'byte',
        },
        eq(settings={}): $.condition.number.length.equal_to(settings=settings),
        equal_to(settings={}): {
          local default = $.condition.number.length.default,

          type: 'number_length_equal_to',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        gt(settings={}): $.condition.number.length.greater_than(settings=settings),
        greater_than(settings={}): {
          local default = $.condition.number.length.default,

          type: 'number_length_greater_than',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        lt(settings={}): $.condition.number.length.less_than(settings=settings),
        less_than(settings={}): {
          local default = $.condition.number.length.default,

          type: 'number_length_less_than',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    meta: {
      condition(settings={}): {
        local default = { condition: null },

        type: 'meta_condition',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      for_each(settings={}): {
        local default = {
          object: $.config.object,
          type: null,
          inspector: null,
        },

        type: 'meta_for_each',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      negate(settings={}): {
        local default = { inspector: null },

        type: 'meta_negate',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    net: $.condition.network,
    network: {
      ip: {
        default: {
          object: $.config.object,
        },
        global_unicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_global_unicast',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        link_local_multicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_link_local_multicast',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        link_local_unicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_link_local_unicast',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        loopback(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_loopback',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        multicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_multicast',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        private(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_private',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        unicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_unicast',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        unspecified(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_unspecified',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        valid(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_valid',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    str: $.condition.string,
    string: {
      default: {
        object: $.config.object,
        value: null,
      },
      has(settings={}): $.condition.string.contains(settings=settings),
      contains(settings={}): {
        local default = $.condition.string.default,

        type: 'string_contains',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      eq(settings={}): $.condition.string.equal_to(settings=settings),
      equal_to(settings={}): {
        local default = $.condition.string.default,

        type: 'string_equal_to',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      gt(settings={}): $.condition.string.greater_than(settings=settings),
      greater_than(settings={}): {
        local default = $.condition.string.default,

        type: 'string_greater_than',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      lt(settings={}): $.condition.string.less_than(settings=settings),
      less_than(settings={}): {
        local default = $.condition.string.default,

        type: 'string_less_than',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      prefix(settings={}): $.condition.string.starts_with(settings=settings),
      starts_with(settings={}): {
        local default = $.condition.string.default,

        type: 'string_starts_with',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      suffix(settings={}): $.condition.string.ends_with(settings=settings),
      ends_with(settings={}): {
        local default = $.condition.string.default,

        type: 'string_ends_with',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      match(settings={}): {
        local default = {
          object: $.config.object,
          pattern: null,
        },

        type: 'string_match',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    util: $.transform.utility,
    utility: {
      random(settings={}): {
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
        arr(settings={}): $.transform.aggregate.from.array(settings=settings),
        array(settings={}): {
          local default = {
            object: $.config.object,
          },

          type: 'aggregate_from_array',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        str(settings={}): $.transform.aggregate.from.string(settings=settings),
        string(settings={}): {
          local default = {
            separator: null,
          },

          type: 'aggregate_from_string',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      to: {
        arr(settings={}): $.transform.aggregate.to.array(settings=settings),
        array(settings={}): {
          local default = {
            object: $.config.object,
            buffer: $.config.buffer,
          },

          type: 'aggregate_to_array',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        str(settings={}): $.transform.aggregate.to.string(settings=settings),
        string(settings={}): {
          local default = {
            buffer: $.config.buffer,
            separator: null,
          },

          type: 'aggregate_to_string',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    arr: $.transform.array,
    array: {
      join(settings={}): {
        local default = {
          object: $.config.object,
          separator: null,
        },

        type: 'array_join',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      zip(settings={}): {
        local default = {
          object: $.config.object,
          as_object: false,
          with_keys: null,
        },

        type: 'array_zip',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    enrich: {
      aws: {
        dynamodb(settings={}): {
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
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        lambda(settings={}): {
          local default = {
            object: $.config.object,
            aws: $.config.aws,
            retry: $.config.retry,
            function_name: null,
          },

          type: 'enrich_aws_lambda',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      dns: {
        default: {
          object: $.config.object,
          request: $.config.request,
        },
        domain_lookup(settings={}): {
          local default = $.transform.enrich.dns.default,

          type: 'enrich_dns_domain_lookup',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        ip_lookup(settings={}): {
          local default = $.transform.enrich.dns.default,

          type: 'enrich_dns_ip_lookup',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        txt_lookup(settings={}): {
          local default = $.transform.enrich.dns.default,

          type: 'enrich_dns_txt_lookup',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      http: {
        default: {
          object: $.config.object,
          request: $.config.request,
          url: null,
          headers: null,
        },
        get(settings={}): {
          local default = $.transform.enrich.http.default,

          type: 'enrich_http_get',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        post(settings={}): {
          local default = $.transform.enrich.http.default { body_key: null },

          type: 'enrich_http_post',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      kv_store: {
        default: {
          object: $.config.object,
          prefix: null,
          kv_store: null,
          close_kv_store: false,
        },
        get(settings={}): {
          local default = $.transform.enrich.kv_store.default,

          type: 'enrich_kv_store_get',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        set(settings={}): {
          local default = $.transform.enrich.kv_store.default { ttl_key: null, ttl_offset: null },

          type: 'enrich_kv_store_set',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    fmt: $.transform.format,
    format: {
      default: {
        object: $.config.object,
      },
      from: {
        b64(settings={}): $.transform.format.from.base64(settings=settings),
        base64(settings={}): {
          local default = $.transform.format.default,

          type: 'format_from_base64',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        gz(settings={}): $.transform.format.from.gzip(settings=settings),
        gzip(settings={}): {
          type: 'format_from_gzip',
        },
        pretty_print(settings={}): {
          type: 'format_from_pretty_print',
        },
      },
      to: {
        b64(settings={}): $.transform.format.to.base64(settings=settings),
        base64(settings={}): {
          local default = $.transform.format.default,

          type: 'format_to_base64',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        gz(settings={}): $.transform.format.to.gzip(settings=settings),
        gzip(settings={}): {
          type: 'format_to_gzip',
        },
      },
    },
    hash: {
      default: {
        object: $.config.object,
      },
      md5(settings={}): {
        local default = $.transform.hash.default,

        type: 'hash_md5',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      sha256(settings={}): {
        local default = $.transform.hash.default,

        type: 'hash_sha256',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    num: $.transform.number,
    number: {
      math: {
        default: {
          object: $.config.object,
        },
        add(settings={}): $.transform.number.math.addition(settings=settings),
        addition(settings={}): {
          local default = $.transform.number.math.default,

          type: 'number_math_addition',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        sub(settings={}): $.transform.number.math.subtraction(settings=settings),
        subtraction(settings={}): {
          local default = $.transform.number.math.default,

          type: 'number_math_subtraction',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        mul(settings={}): $.transform.number.math.multiplication(settings=settings),
        multiplication(settings={}): {
          local default = $.transform.number.math.default,

          type: 'number_math_multiplication',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        div(settings={}): $.transform.number.math.division(settings=settings),
        division(settings={}): {
          local default = $.transform.number.math.default,

          type: 'number_math_division',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    meta: {
      err(settings={}): {
        local default = { transform: null },

        type: 'meta_err',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      for_each(settings={}): {
        local default = {
          object: $.config.object,
          transform: null,
        },

        type: 'meta_for_each',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      metric: {
        duration(settings={}): {
          local default = {
            metric: $.config.metric,
            transform: null,
          },

          type: 'meta_metric_duration',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      pipe(settings={}): $.transform.meta.pipeline(settings=settings),
      pipeline(settings={}): {
        local default = {
          object: $.config.object,
          transform: null,
        },

        type: 'meta_pipeline',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      switch(settings={}): {
        local default = { cases: null },

        type: 'meta_switch',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    net: $.transform.network,
    network: {
      domain: {
        default: {
          object: $.config.object,
        },
        registered_domain(settings={}): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_registered_domain',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        subdomain(settings={}): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_subdomain',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        tld(settings={}): $.transform.network.domain.top_level_domain(settings=settings),
        top_level_domain(settings={}): {
          local default = $.transform.network.domain.default,

          type: 'network_domain_top_level_domain',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    obj: $.transform.object,
    object: {
      default: {
        object: $.config.object,
      },
      cp(settings={}): $.transform.object.copy(settings=settings),
      copy(settings={}): {
        local default = $.transform.object.default,

        type: 'object_copy',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      del(settings={}): $.transform.object.delete(settings=settings),
      delete(settings={}): {
        local default = $.transform.object.default,

        type: 'object_delete',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      insert(settings={}): {
        local default = $.transform.object.default,

        type: 'object_insert',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      jq(settings={}): {
        local default = { filter: null },

        type: 'object_jq',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      to: {
        bool(settings={}): $.transform.object.to.boolean(settings=settings),
        boolean(settings={}): {
          local default = $.transform.object.default,

          type: 'object_to_boolean',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        float(settings={}): {
          local default = $.transform.object.default,

          type: 'object_to_float',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        int(settings={}): $.transform.object.to.integer(settings=settings),
        integer(settings={}): {
          local default = $.transform.object.default,

          type: 'object_to_integer',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        str(settings={}): $.transform.object.to.string(settings=settings),
        string(settings={}): {
          local default = $.transform.object.default,

          type: 'object_to_string',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        uint(settings={}): $.transform.object.to.unsigned_integer(settings=settings),
        unsigned_integer(settings={}): {
          local default = $.transform.object.default,

          type: 'object_to_unsigned_integer',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
    },
    send: {
      aws: {
        dynamodb(settings={}): {
          local default = {
            aws: $.config.aws,
            retry: $.config.retry,
            table_name: null,
          },

          type: 'send_aws_dynamodb',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        firehose(settings={}): $.transform.send.aws.kinesis_data_firehose(settings=settings),
        kinesis_data_firehose(settings={}): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            stream_name: null,
          },

          type: 'send_aws_kinesis_data_firehose',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        kinesis_data_stream(settings={}): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            stream_name: null,
            partition: null,
            partition_key: null,
            aggregation: false,
          },

          type: 'send_aws_kinesis_data_stream',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        s3(settings={}): {
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
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        sns(settings={}): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            arn: null,
          },

          type: 'send_aws_sns',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        sqs(settings={}): {
          local default = {
            aws: $.config.aws,
            buffer: $.config.buffer,
            retry: $.config.retry,
            arn: null,
          },

          type: 'send_aws_sqs',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      file(settings={}): {
        local default = {
          buffer: $.config.buffer,
          file_path: $.file_path,
          file_format: { type: 'json' },
          file_compression: { type: 'gzip' },
        },

        type: 'send_file',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      http: {
        post(settings={}): {
          local default = {
            url: null,
            headers: null,
            headers_key: null,
          },

          local s = std.mergePatch(settings, {
            headers: if std.objectHas(settings, 'headers') then settings.headers else if std.objectHas(settings, 'hdr') then settings.hdr else null,
            headers_key: if std.objectHas(settings, 'headers_key') then settings.headers_key else if std.objectHas(settings, 'hdr_key') then settings.hdr_key else null,
            hdr: null,
            hdr_key: null,
          }),

          type: 'send_http_post',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(s))),
        },
      },
      stdout(settings={}): {
        type: 'send_stdout',
      },
    },
    str: $.transform.string,
    string: {
      append(settings={}): {
        local default = {
          object: $.config.object,
          suffix: null,
        },

        type: 'string_append',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      capture(settings={}): {
        local default = {
          object: $.config.object,
          pattern: null,
          count: 0,
        },

        type: 'string_capture',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      repl: $.transform.string.replace,
      replace(settings={}): {
        local default = {
          object: $.config.object,
          pattern: null,
          replacement: null,
        },

        local s = std.mergePatch(settings, {
          pattern: settings.pattern,
          replacement: if std.objectHas(settings, 'replacement') then settings.replacement else if std.objectHas(settings, 'repl') then settings.repl else null,
          repl: null,
        }),

        type: 'string_replace',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(s))),
      },
      split(settings={}): {
        local default = {
          object: $.config.object,
          separator: null,
        },

        local s = std.mergePatch(settings, {
          separator: if std.objectHas(settings, 'separator') then settings.separator else if std.objectHas(settings, 'sep') then settings.sep else null,
          sep: null,
        }),

        type: 'string_split',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(s))),
      },
      to: {
        default: {
          object: $.config.object,
        },
        lower(settings={}): {
          local default = $.transform.string.to.default,

          type: 'string_to_lower',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        upper(settings={}): {
          local default = $.transform.string.to.default,

          type: 'string_to_upper',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        snake(settings={}): {
          local default = $.transform.string.to.default,

          type: 'string_to_snake',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      uuid(settings={}): {
        local default = {
          object: $.config.object,
        },

        type: 'string_uuid',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    time: {
      from: {
        str(settings={}): $.transform.time.from.string(settings=settings),
        string(settings={}): {
          local default = {
            object: $.config.object,
            format: null,
            location: null,
          },

          type: 'time_from_string',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
        unix(settings={}): {
          local default = {
            object: $.config.object,
          },

          type: 'time_from_unix',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      now(settings={}): {
        local default = {
          object: $.config.object,
        },

        type: 'time_now',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      to: {
        str(settings={}): $.transform.time.to.string(settings=settings),
        string(settings={}): {
          local default = {
            object: $.config.object,
            format: null,
            location: null,
          },

          type: 'time_to_string',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      unix(settings={}): {
        local default = {
          object: $.config.object,
        },

        type: 'time_to_unix',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
    util: $.transform.utility,
    utility: {
      delay(settings={}): {
        local default = {
          duration: null,
        },

        type: 'utility_delay',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      drop(settings={}): {
        type: 'utility_drop',
      },
      err(settings={}): {
        local default = {
          message: null,
        },

        type: 'utility_err',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
      metric: {
        count(settings={}): {
          local default = {
            metric: $.config.metric,
          },

          type: 'utility_metric_count',
          settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
        },
      },
      secret(settings={}): {
        local default = { secret: null },

        type: 'utility_secret',
        settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
      },
    },
  },
  // Mirrors interfaces from the internal/kv_store package.
  kv_store: {
    aws_dynamodb(settings={}): {
      local default = {
        aws: $.config.aws,
        retry: $.config.retry,
        table_name: null,
        attributes: { partition_key: null, sort_key: null, value: null, ttl: null },
        consistent_read: false,
      },

      type: 'aws_dynamodb',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
    csv_file(settings={}): {
      local default = { file: null, column: null, delimiter: ',', header: null },

      type: 'csv_file',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
    json_file(settings=$.defaults.kv_store.json_file.settings): {
      local default = { file: null, is_lines: false },

      type: 'json_file',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
    memory(settings={}): {
      local default = { capacity: 1024 },

      type: 'memory',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
    mmdb(settings={}): {
      local default = { file: null },

      type: 'mmdb',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
    text_file(settings={}): {
      local default = { file: null },

      type: 'text_file',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
  },
  // Mirrors structs from the internal/config package.
  config: {
    aws: { region: null, role_arn: null },
    buffer: { count: 1000, size: 100000, duration: '5m', key: null },
    metric: { name: null, attributes: null, destination: null },
    object: { src_key: null, dst_key: null },
    request: { timeout: '1s' },
    retry: { count: 3 },
  },
  // Mirrors config from the internal/file package.
  file_path: { prefix: null, prefix_key: null, time_format: '2006/01/02', uuid: true, extension: true },
  // Mirrors interfaces from the internal/secrets package.
  secrets: {
    default: { id: null, ttl: null },
    aws_secrets_manager(settings={}): {
      local default = {
        id: null,
        name: null,
        ttl_offset: null,
        aws: $.config.aws,
        retry: $.config.retry,
      },

      type: 'aws_secrets_manager',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
    environment_variable(settings={}): {
      local default = { id: null, name: null, ttl_offset: null },

      type: 'environment_variable',
      settings: std.prune(std.mergePatch(default, $.helpers.abbv(settings))),
    },
  },
  // Commonly used condition and transform patterns.
  pattern: {
    cnd: $.pattern.condition,
    condition: {
      obj(key): {
        object: { source: key },
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
            $.condition.number.length.equal_to(settings=$.pattern.condition.obj(key) { value: 0 }),
          // Checks if data is greater than zero.
          //
          // Use with the ANY / ALL operator to match non-empty data.
          // Use with the NONE operator to match empty data.
          gt_zero(key=null):
            $.condition.number.length.greater_than(settings=$.pattern.condition.obj(key) { value: 0 }),
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
        settings: { cases: [{ condition: c, transform: transform }] },
      },
    },
  },
  // Utility functions that can be used in conditions and transforms.
  helpers: {
    // If the input is not an array, then this returns it as an array.
    make_array(i): if !std.isArray(i) then [i] else i,
    obj: $.helpers.object,
    object: {
      // If key is `foo` and arr is `bar`, then the result is `foo.bar`.
      // If key is `foo` and arr is `[bar, baz]`, then the result is `foo.bar.baz`.
      append(key, arr): std.join('.', $.helpers.make_array(key) + $.helpers.make_array(arr)),
      // If key is `foo`, then the result is `foo.-1`.
      append_array(key): key + '.-1',
      // If key is `foo` and e is `0`, then the result is `foo.0`.
      get_element(key, e=0): std.join('.', [key, if std.isNumber(e) then std.toString(e) else e]),
    },
    abbv(settings): std.mergePatch(settings, {
      object: if std.objectHas(settings, 'object') then $.helpers.abbv_obj(settings.object) else if std.objectHas(settings, 'obj') then $.helpers.abbv_obj(settings.obj) else null,
      obj: null,
    }),
    abbv_obj(s): {
      src_key: if std.objectHas(s, 'src') then s.src else if std.objectHas(s, 'src_key') then s.src_key else null,
      dst_key: if std.objectHas(s, 'dst') then s.dst else if std.objectHas(s, 'dst_key') then s.dst_key else null,
      src: null,
      dst: null,
    },
  },
}
