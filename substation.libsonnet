local helpers = {
  // If the input is not an array, then this returns it as an array.
  make_array(i): if !std.isArray(i) then [i] else i,
  abbv(settings): std.mergePatch(settings, {
    object: if std.objectHas(settings, 'object') then $.abbv_obj(settings.object) else if std.objectHas(settings, 'obj') then $.abbv_obj(settings.obj) else null,
    obj: null,
  }),
  abbv_obj(s): {
    source_key: if std.objectHas(s, 'src') then s.src else if std.objectHas(s, 'source_key') then s.source_key else null,
    src: null,
    target_key: if std.objectHas(s, 'trg') then s.trg else if std.objectHas(s, 'target_key') then s.target_key else null,
    trg: null,
    batch_key: if std.objectHas(s, 'btch') then s.batch else if std.objectHas(s, 'batch_key') then s.batch_key else null,
  },
  id(type, settings): std.join('-', [std.md5(type)[:8], std.md5(std.toString(settings))[:8]]),
};

{
  // Mirrors interfaces from the condition package.
  cnd: $.condition,
  condition: {
    all(i): $.condition.meta.all({ conditions: helpers.make_array(i) }),
    any(i): $.condition.meta.any({ conditions: helpers.make_array(i) }),
    none(i): $.condition.meta.none({ conditions: helpers.make_array(i) }),
    meta: {
      all(settings={}): {
        local default = {
          object: $.config.object,
          conditions: [],
        },

        type: 'meta_all',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      any(settings={}): {
        local default = {
          object: $.config.object,
          conditions: [],
        },

        type: 'meta_any',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      none(settings={}): {
        local default = {
          object: $.config.object,
          conditions: [],
        },

        type: 'meta_none',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
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
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    num: $.condition.number,
    number: {
      default: {
        object: $.config.object,
        value: null,
      },
      eq(settings={}): $.condition.number.equal_to(settings=settings),
      equal_to(settings={}): {
        local default = $.condition.number.default,

        type: 'number_equal_to',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      lt(settings={}): $.condition.number.less_than(settings=settings),
      less_than(settings={}): {
        local default = $.condition.number.default,

        type: 'number_less_than',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      gt(settings={}): $.condition.number.greater_than(settings=settings),
      greater_than(settings={}): {
        local default = $.condition.number.default,

        type: 'number_greater_than',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      bitwise: {
        and(settings={}): {
          local default = {
            object: $.config.object,
            value: null,
          },

          type: 'number_bitwise_and',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        not(settings={}): {
          local default = {
            object: $.config.object,
          },

          type: 'number_bitwise_not',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        or(settings={}): {
          local default = {
            object: $.config.object,
            value: null,
          },

          type: 'number_bitwise_or',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        xor(settings={}): {
          local default = {
            object: $.config.object,
            value: null,
          },

          type: 'number_bitwise_xor',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
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
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        gt(settings={}): $.condition.number.length.greater_than(settings=settings),
        greater_than(settings={}): {
          local default = $.condition.number.length.default,

          type: 'number_length_greater_than',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        lt(settings={}): $.condition.number.length.less_than(settings=settings),
        less_than(settings={}): {
          local default = $.condition.number.length.default,

          type: 'number_length_less_than',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
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
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        link_local_multicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_link_local_multicast',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        link_local_unicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_link_local_unicast',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        loopback(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_loopback',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        multicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_multicast',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        private(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_private',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        unicast(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_unicast',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        unspecified(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_unspecified',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        valid(settings={}): {
          local default = $.condition.network.ip.default,

          type: 'network_ip_valid',
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
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
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      eq(settings={}): $.condition.string.equal_to(settings=settings),
      equal_to(settings={}): {
        local default = $.condition.string.default,

        type: 'string_equal_to',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      gt(settings={}): $.condition.string.greater_than(settings=settings),
      greater_than(settings={}): {
        local default = $.condition.string.default,

        type: 'string_greater_than',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      lt(settings={}): $.condition.string.less_than(settings=settings),
      less_than(settings={}): {
        local default = $.condition.string.default,

        type: 'string_less_than',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      prefix(settings={}): $.condition.string.starts_with(settings=settings),
      starts_with(settings={}): {
        local default = $.condition.string.default,

        type: 'string_starts_with',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      suffix(settings={}): $.condition.string.ends_with(settings=settings),
      ends_with(settings={}): {
        local default = $.condition.string.default,

        type: 'string_ends_with',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      match(settings={}): {
        local default = {
          object: $.config.object,
          pattern: null,
        },

        type: 'string_match',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    util: $.condition.utility,
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
          local type = 'aggregate_from_array',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        str(settings={}): $.transform.aggregate.from.string(settings=settings),
        string(settings={}): {
          local type = 'aggregate_from_string',
          local default = {
            id: helpers.id(type, settings),
            separator: null,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      to: {
        arr(settings={}): $.transform.aggregate.to.array(settings=settings),
        array(settings={}): {
          local type = 'aggregate_to_array',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
            batch: $.config.batch,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        str(settings={}): $.transform.aggregate.to.string(settings=settings),
        string(settings={}): {
          local type = 'aggregate_to_string',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            separator: null,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
    },
    arr: $.transform.array,
    array: {
      join(settings={}): {
        local type = 'array_join',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          separator: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      zip(settings={}): {
        local type = 'array_zip',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    enrich: {
      aws: {
        dynamodb: {
          query(settings={}): {
            local type = 'enrich_aws_dynamodb_query',
            local default = {
              id: helpers.id(type, settings),
              object: $.config.object,
              aws: $.config.aws,
              attributes: { partition_key: null, sort_key: null },
              limit: 1,
              scan_index_forward: false,
            },

            type: type,
            settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
          },
        },
        lambda(settings={}): {
          local type = 'enrich_aws_lambda',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
            aws: $.config.aws,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      dns: {
        default: {
          object: $.config.object,
          request: $.config.request,
        },
        domain_lookup(settings={}): {
          local type = 'enrich_dns_domain_lookup',
          local default = $.transform.enrich.dns.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        ip_lookup(settings={}): {
          local type = 'enrich_dns_ip_lookup',
          local default = $.transform.enrich.dns.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        txt_lookup(settings={}): {
          local type = 'enrich_dns_txt_lookup',
          local default = $.transform.enrich.dns.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
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
          local type = 'enrich_http_get',
          local default = $.transform.enrich.http.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        post(settings={}): {
          local type = 'enrich_http_post',
          local default = $.transform.enrich.http.default { body_key: null, id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      kv_store: {
        default: {
          object: $.config.object,
          prefix: null,
          kv_store: null,
          close_kv_store: false,
        },
        iget: $.transform.enrich.kv_store.item.get,
        iset: $.transform.enrich.kv_store.item.set,
        item: {
          get(settings={}): {
            local type = 'enrich_kv_store_get',
            local default = $.transform.enrich.kv_store.default { id: helpers.id(type, settings) },

            type: type,
            settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
          },
          set(settings={}): {
            local type = 'enrich_kv_store_set',
            local default = $.transform.enrich.kv_store.default { ttl_key: null, ttl_offset: '0s', id: helpers.id(type, settings) },

            type: type,
            settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
          },
        },
        sadd: $.transform.enrich.kv_store.set.add,
        set: {
          add(settings={}): {
            local type = 'enrich_kv_store_set_add',
            local default = $.transform.enrich.kv_store.default { ttl_key: null, ttl_offset: '0s', id: helpers.id(type, settings) },

            type: type,
            settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
          },
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
          local type = 'format_from_base64',
          local default = $.transform.format.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        gz(settings={}): $.transform.format.from.gzip(settings=settings),
        gzip(settings={}): {
          local type = 'format_from_gzip',
          local default = { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        pretty_print(settings={}): {
          local type = 'format_from_pretty_print',
          local default = { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        zip(settings={}): {
          local type = 'format_from_zip',
          local default = { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      to: {
        b64(settings={}): $.transform.format.to.base64(settings=settings),
        base64(settings={}): {
          local type = 'format_to_base64',
          local default = $.transform.format.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        gz(settings={}): $.transform.format.to.gzip(settings=settings),
        gzip(settings={}): {
          local type = 'format_to_gzip',
          local default = { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
    },
    hash: {
      default: {
        object: $.config.object,
      },
      md5(settings={}): {
        local type = 'hash_md5',
        local default = $.transform.hash.default { id: helpers.id(type, settings) },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      sha256(settings={}): {
        local type = 'hash_sha256',
        local default = $.transform.hash.default { id: helpers.id(type, settings) },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    num: $.transform.number,
    number: {
      max(settings={}): $.transform.number.maximum(settings=settings),
      maximum(settings={}): {
        local type = 'number_maximum',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          value: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      min(settings={}): $.transform.number.minimum(settings=settings),
      minimum(settings={}): {
        local type = 'number_minimum',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          value: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      math: {
        default: {
          object: $.config.object,
        },
        add(settings={}): $.transform.number.math.addition(settings=settings),
        addition(settings={}): {
          local type = 'number_math_addition',
          local default = $.transform.number.math.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        sub(settings={}): $.transform.number.math.subtraction(settings=settings),
        subtraction(settings={}): {
          local type = 'number_math_subtraction',
          local default = $.transform.number.math.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        mul(settings={}): $.transform.number.math.multiplication(settings=settings),
        multiplication(settings={}): {
          local type = 'number_math_multiplication',
          local default = $.transform.number.math.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        div(settings={}): $.transform.number.math.division(settings=settings),
        division(settings={}): {
          local type = 'number_math_division',
          local default = $.transform.number.math.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
    },
    meta: {
      err(settings={}): {
        local type = 'meta_err',
        local default = {
          id: helpers.id(type, settings),
          transforms: null,
          error_messages: ['.*'],
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      for_each(settings={}): {
        local type = 'meta_for_each',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          transforms: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      kv_store: {
        lock(settings={}): {
          local type = 'meta_kv_store_lock',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object { ttl_key: null },
            transforms: null,
            kv_store: null,
            prefix: null,
            ttl_offset: '0s',
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      metric: {
        duration(settings={}): {
          local type = 'meta_metric_duration',
          local default = {
            id: helpers.id(type, settings),
            metric: $.config.metric,
            transforms: null,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      retry(settings={}): {
        local type = 'meta_retry',
        local default = {
          id: helpers.id(type, settings),
          retry: $.config.retry,
          transforms: null,
          condition: null,
          error_messages: ['.*'],
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      switch(settings={}): {
        local type = 'meta_switch',
        local default = {
          id: helpers.id(type, settings),
          cases: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    net: $.transform.network,
    network: {
      domain: {
        default: {
          object: $.config.object,
        },
        registered_domain(settings={}): {
          local type = 'network_domain_registered_domain',
          local default = $.transform.network.domain.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        subdomain(settings={}): {
          local type = 'network_domain_subdomain',
          local default = $.transform.network.domain.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        tld(settings={}): $.transform.network.domain.top_level_domain(settings=settings),
        top_level_domain(settings={}): {
          local type = 'network_domain_top_level_domain',
          local default = $.transform.network.domain.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
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
        local type = 'object_copy',
        local default = $.transform.object.default { id: helpers.id(type, settings) },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      del(settings={}): $.transform.object.delete(settings=settings),
      delete(settings={}): {
        local type = 'object_delete',
        local default = $.transform.object.default { id: helpers.id(type, settings) },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      insert(settings={}): {
        local type = 'object_insert',
        local default = $.transform.object.default { id: helpers.id(type, settings) },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      jq(settings={}): {
        local type = 'object_jq',
        local default = {
          id: helpers.id(type, settings),
          filter: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      to: {
        bool(settings={}): $.transform.object.to.boolean(settings=settings),
        boolean(settings={}): {
          local type = 'object_to_boolean',
          local default = $.transform.object.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        float(settings={}): {
          local type = 'object_to_float',
          local default = $.transform.object.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        int(settings={}): $.transform.object.to.integer(settings=settings),
        integer(settings={}): {
          local type = 'object_to_integer',
          local default = $.transform.object.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        str(settings={}): $.transform.object.to.string(settings=settings),
        string(settings={}): {
          local type = 'object_to_string',
          local default = $.transform.object.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        uint(settings={}): $.transform.object.to.unsigned_integer(settings=settings),
        unsigned_integer(settings={}): {
          local type = 'object_to_unsigned_integer',
          local default = $.transform.object.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
    },
    send: {
      aws: {
        dynamodb: {
          put(settings={}): {
            local type = 'send_aws_dynamodb_put',
            local default = {
              id: helpers.id(type, settings),
              batch: $.config.batch,
              aws: $.config.aws,
              auxiliary_transforms: null,
            },

            local s = std.mergePatch(settings, {
              auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
              aux_tforms: null,
            }),

            type: type,
            settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
          },
        },
        firehose(settings={}): $.transform.send.aws.data_firehose(settings=settings),
        data_firehose(settings={}): {
          local type = 'send_aws_data_firehose',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            auxiliary_transforms: null,
          },

          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
        eventbridge(settings={}): {
          local type = 'send_aws_eventbridge',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            auxiliary_transforms: null,
            description: null,
          },
          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
        kinesis_data_stream(settings={}): {
          local type = 'send_aws_kinesis_data_stream',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            auxiliary_transforms: null,
            use_batch_key_as_partition_key: false,
            enable_record_aggregation: false,
          },

          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
        lambda(settings={}): {
          local type = 'send_aws_lambda',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            auxiliary_transforms: null,
          },

          type: type,
          settings: std.mergePatch(default, helpers.abbv(settings)),
        },
        s3(settings={}): {
          local type = 'send_aws_s3',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            file_path: $.file_path,
            auxiliary_transforms: null,
            storage_class: 'STANDARD',
          },

          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
        sns(settings={}): {
          local type = 'send_aws_sns',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            auxiliary_transforms: null,
          },

          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
        sqs(settings={}): {
          local type = 'send_aws_sqs',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            aws: $.config.aws,
            auxiliary_transforms: null,
          },

          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
      },
      file(settings={}): {
        local type = 'send_file',
        local default = {
          id: helpers.id(type, settings),
          batch: $.config.batch,
          auxiliary_transforms: null,
          file_path: $.file_path,
        },

        local s = std.mergePatch(settings, {
          auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
          aux_tforms: null,
        }),

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
      },
      http: {
        post(settings={}): {
          local type = 'send_http_post',
          local default = {
            id: helpers.id(type, settings),
            batch: $.config.batch,
            auxiliary_transforms: null,
            url: null,
            headers: null,
          },

          local s = std.mergePatch(settings, {
            auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
            aux_tforms: null,
            headers: if std.objectHas(settings, 'headers') then settings.headers else if std.objectHas(settings, 'hdr') then settings.hdr else null,
            hdr: null,
          }),

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
        },
      },
      stdout(settings={}): {
        local type = 'send_stdout',
        local default = {
          id: helpers.id(type, settings),
          batch: $.config.batch,
          auxiliary_transforms: null,
        },

        local s = std.mergePatch(settings, {
          auxiliary_transforms: if std.objectHas(settings, 'auxiliary_transforms') then settings.auxiliary_transforms else if std.objectHas(settings, 'aux_tforms') then settings.aux_tforms else null,
          aux_tforms: null,
        }),

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
      },
    },
    str: $.transform.string,
    string: {
      append(settings={}): {
        local type = 'string_append',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          suffix: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      capture(settings={}): {
        local type = 'string_capture',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          pattern: null,
          count: 0,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      repl: $.transform.string.replace,
      replace(settings={}): {
        local type = 'string_replace',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          pattern: null,
          replacement: null,
        },

        local s = std.mergePatch(settings, {
          pattern: settings.pattern,
          replacement: if std.objectHas(settings, 'replacement') then settings.replacement else if std.objectHas(settings, 'repl') then settings.repl else null,
          repl: null,
        }),

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
      },
      split(settings={}): {
        local type = 'string_split',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
          separator: null,
        },

        local s = std.mergePatch(settings, {
          separator: if std.objectHas(settings, 'separator') then settings.separator else if std.objectHas(settings, 'sep') then settings.sep else null,
          sep: null,
        }),

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(s))),
      },
      to: {
        default: {
          object: $.config.object,
        },
        lower(settings={}): {
          local type = 'string_to_lower',
          local default = $.transform.string.to.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        upper(settings={}): {
          local type = 'string_to_upper',
          local default = $.transform.string.to.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        snake(settings={}): {
          local type = 'string_to_snake',
          local default = $.transform.string.to.default { id: helpers.id(type, settings) },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      uuid(settings={}): {
        local type = 'string_uuid',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    time: {
      from: {
        str(settings={}): $.transform.time.from.string(settings=settings),
        string(settings={}): {
          local type = 'time_from_string',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
            format: null,
            location: 'UTC',
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        unix(settings={}): {
          local type = 'time_from_unix',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        unix_milli(settings={}): {
          local type = 'time_from_unix_milli',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      now(settings={}): {
        local type = 'time_now',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      to: {
        str(settings={}): $.transform.time.to.string(settings=settings),
        string(settings={}): {
          local type = 'time_to_string',
          local default = {
            id: helpers.id(type, settings),
            object: $.config.object,
            format: null,
            location: 'UTC',
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      unix(settings={}): {
        local type = 'time_to_unix',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      unix_milli(settings={}): {
        local type = 'time_to_unix_milli',
        local default = {
          id: helpers.id(type, settings),
          object: $.config.object,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    util: $.transform.utility,
    utility: {
      control(settings={}): {
        local type = 'utility_control',
        local default = {
          id: helpers.id(type, settings),
          batch: $.config.batch,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      delay(settings={}): {
        local type = 'utility_delay',
        local default = {
          id: helpers.id(type, settings),
          duration: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      drop(settings={}): {
        local type = 'utility_drop',
        local default = {
          id: helpers.id(type, settings),
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      err(settings={}): {
        local type = 'utility_err',
        local default = {
          id: helpers.id(type, settings),
          message: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      message(settings={}): {
        local type = 'utility_message',
        local default = {
          id: helpers.id(type, settings),
          value: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
      metric: {
        bytes(settings={}): {
          local type = 'utility_metric_bytes',
          local default = {
            id: helpers.id(type, settings),
            metric: $.config.metric,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        count(settings={}): {
          local type = 'utility_metric_count',
          local default = {
            id: helpers.id(type, settings),
            metric: $.config.metric,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
        freshness(settings={}): {
          local type = 'utility_metric_freshness',
          local default = {
            id: helpers.id(type, settings),
            threshold: null,
            metric: $.config.metric,
            object: $.config.object,
          },

          type: type,
          settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
        },
      },
      secret(settings={}): {
        local type = 'utility_secret',
        local default = {
          id: helpers.id(type, settings),
          secret: null,
        },

        type: type,
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
  },
  // Mirrors interfaces from the internal/kv_store package.
  kv_store: {
    aws: {
      dynamodb(settings={}): {
        local default = {
          aws: $.config.aws,
          attributes: { partition_key: null, sort_key: null, value: null, ttl: null },
          consistent_read: false,
        },

        type: 'aws_dynamodb',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    csv_file(settings={}): {
      local default = { file: null, column: null, delimiter: ',', header: null },

      type: 'csv_file',
      settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
    },
    json_file(settings=$.defaults.kv_store.json_file.settings): {
      local default = { file: null, is_lines: false },

      type: 'json_file',
      settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
    },
    memory(settings={}): {
      local default = { capacity: 1024 },

      type: 'memory',
      settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
    },
    mmdb(settings={}): {
      local default = { file: null },

      type: 'mmdb',
      settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
    },
    text_file(settings={}): {
      local default = { file: null },

      type: 'text_file',
      settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
    },
  },
  // Mirrors interfaces from the internal/secrets package.
  secrets: {
    default: { id: null, ttl: null },
    aws: {
      secrets_manager(settings={}): {
        local default = {
          aws: $.config.aws,
          id: null,
          ttl_offset: null,
        },

        type: 'aws_secrets_manager',
        settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
      },
    },
    environment_variable(settings={}): {
      local default = { id: null, name: null, ttl_offset: null },

      type: 'environment_variable',
      settings: std.prune(std.mergePatch(default, helpers.abbv(settings))),
    },
  },
  // Mirrors structs from the internal/config package.
  config: {
    aws: { arn: null, assume_role_arn: null },
    batch: { count: 1000, size: 1000 * 1000, duration: '1m' },
    metric: { name: null, attributes: null, destination: null },
    object: { source_key: null, target_key: null, batch_key: null },
    request: { timeout: '1s' },
    retry: { count: 3, delay: '1s' },
  },
  // Mirrors config from the internal/file package.
  file_path: { prefix: null, time_format: '2006/01/02', uuid: true, suffix: null },
}
