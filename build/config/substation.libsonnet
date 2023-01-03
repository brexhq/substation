{
  defaults: {
    processor: {
      time: {
        set_format: '2006-01-02T15:04:05.000000Z',
      }
    }
  },
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
    inspector: {
      // validates base settings of any inspector by checking for the
      // existence of any fields except key and negate
      validate(settings): std.all([
        if !std.member(['key', 'negate'], x) then false else true
        for x in std.objectFields(settings)
      ]),
    },
    // dynamically flattens processor configurations
    flatten_processors(processor): std.flattenArrays([
      if std.objectHas(p, 'processor') then
        if std.isArray(p.processor) then p.processor
        else [p.processor]
      // else if std.objectHas(p, 'processors') then
      //   if std.isArray(p.processors) then p.processor
      //   else [p.processors]
      else [p]

      for p in $.helpers.make_array(processor)
    ]),
  },
  interfaces: {
    // mirrors interfaces from the condition package
    operator: {
      all(i): { operator: 'all', inspectors: if !std.isArray(i) then [i] else i },
      any(i): { operator: 'any', inspectors: if !std.isArray(i) then [i] else i },
      none(i): { operator: 'none', inspectors: if !std.isArray(i) then [i] else i },
    },
    inspector: {
      settings: { key: null, negate: null },
      content(type, settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'content',
        settings: std.mergePatch({
          options: {
            type: type,
          },
        }, s),
      },
      for_each(type, inspector, settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'for_each',
        settings: std.mergePatch({
          options: {
            type: type,
            inspector: inspector,
          },
        }, s),
      },
      ip(type, settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'ip',
        settings: std.mergePatch({
          options: {
            type: type,
          },
        }, s),
      },
      json_schema(schema, settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'json_schema',
        settings: std.mergePatch({
          options: {
            schema: schema,
          },
        }, s),
      },
      json_valid(settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'json_valid',
        settings: s,
      },
      length(type, value, measurement='bytes', settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'length',
        settings: std.mergePatch({
          options: {
            type: type,
            value: value,
            measurement: measurement,
          },
        }, s),
      },
      random: {
        type: 'random',
      },
      regexp(expression, settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'regexp',
        settings: std.mergePatch({
          options: {
            expression: expression,
          },
        }, s),
      },
      strings(type, expression, settings=$.interfaces.inspector.settings): {
        assert $.helpers.inspector.validate(settings) : 'invalid inspector settings',
        local s = std.mergePatch($.interfaces.inspector.settings, settings),

        type: 'strings',
        settings: std.mergePatch({
          options: {
            type: type,
            expression: expression,
          },
        }, s),
      },
    },
    // mirrors interfaces from the process package
    processor: {
      settings: { key: null, set_key: null, condition: null, ignore_close: null, ignore_errors: null },
      aggregate(key=null,
                separator=null,
                max_count=1000,
                max_size=10000,
                settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'aggregate',
        settings: std.mergePatch({
          options: { key: key, separator: separator, max_count: max_count, max_size: max_size },
        }, s),
      },
      aws_dynamodb(table,
                   key_condition_expression,
                   limit=1,
                   scan_index_forward=false,
                   settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'aws_dynamodb',
        settings: std.mergePatch({
          options: {
            table: table,
            key_condition_expression: key_condition_expression,
            limit: limit,
            scan_index_forward: scan_index_forward,
          },
        }, s),
      },
      aws_lambda(function_name,
                 settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'aws_lambda',
        settings: std.mergePatch({
          options: { function_name: function_name },
        }, s),
      },
      base64(direction,
             settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'base64',
        settings: std.mergePatch({
          options: { direction: direction },
        }, s),
      },
      capture(expression,
              type='find',
              count=-1,
              settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'capture',
        settings: std.mergePatch({
          options: {
            expression: expression,
            type: type,
            count: count,
          },
        }, s),
      },
      case(type,
           settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'case',
        settings: std.mergePatch({
          options: { type: type },
        }, s),
      },
      convert(type,
              settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'convert',
        settings: std.mergePatch({
          options: { type: type },
        }, s),
      },
      copy(settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'copy',
        settings: s,
      },
      delete(settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'delete',
        settings: s,
      },
      dns(type,
          timeout=1000,
          settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'dns',
        settings: std.mergePatch({
          options: { type: type, timeout: timeout },
        }, s),
      },
      domain(type,
             settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'domain',
        settings: std.mergePatch({
          options: { type: type },
        }, s),
      },
      drop(settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'drop',
        settings: settings,
      },
      expand(settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'expand',
        settings: settings,
      },
      flatten(deep=true,
              settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'flatten',
        settings: std.mergePatch({
          options: { deep: deep },
        }, s),
      },
      for_each(processor,
               settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'for_each',
        settings: std.mergePatch({
          options: { processor: processor },
        }, s),
      },
      group(keys=[],
            settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'group',
        settings: std.mergePatch({
          options: { keys: keys },
        }, s),
      },
      gzip(direction,
           settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'gzip',
        settings: std.mergePatch({
          options: { direction: direction },
        }, s),
      },
      hash(algorithm='sha256',
           settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'hash',
        settings: std.mergePatch({
          options: { algorithm: algorithm },
        }, s),
      },
      insert(value,
             settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'insert',
        settings: std.mergePatch({
          options: { value: value },
        }, s),
      },
      ip_database(options,
                  settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'ip_database',
        settings: std.mergePatch({
          options: options,
        }, s),
      },
      join(separator,
           settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'join',
        settings: std.mergePatch({
          options: { separator: separator },
        }, s),
      },
      math(operation,
           settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'math',
        settings: std.mergePatch({
          options: { operation: operation },
        }, s),
      },
      pipeline(processors,
               settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'pipeline',
        settings: std.mergePatch({
          options: { processors: processors },
        }, s),
      },
      pretty_print(direction,
                   settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'pretty_print',
        settings: std.mergePatch({
          options: { direction: direction },
        }, s),
      },
      replace(old,
              new,
              count=-1,
              settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'replace',
        settings: std.mergePatch({
          options: { old: old, new: new, count: count },
        }, s),
      },
      split(separator,
            settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'split',
        settings: std.mergePatch({
          options: { separator: separator },
        }, s),
      },
      time(format,
           location=null,
           set_format=$.defaults.processor.time.set_format,
           set_location=null,
           settings=$.interfaces.processor.settings): {
        local s = std.mergePatch($.interfaces.processor.settings, settings),

        type: 'time',
        settings: std.mergePatch({
          options: { format: format, location: location, set_format: set_format, set_location: set_location },
        }, s),
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
  },
  patterns: {
    inspector: {
      // negates any inspector
      negate(inspector): std.mergePatch(inspector, { settings: { negate: true } }),
      ip: {
        // checks if an IP address is private.
        //
        // use with the ANY operator to match private IP addresses.
        // use with the NONE operator to match public IP addresses.
        private(key=null): [
          $.interfaces.inspector.ip(type='loopback', settings={ key: key }),
          $.interfaces.inspector.ip(type='multicast', settings={ key: key }),
          $.interfaces.inspector.ip(type='multicast_link_local', settings={ key: key }),
          $.interfaces.inspector.ip(type='private', settings={ key: key }),
          $.interfaces.inspector.ip(type='unicast_link_local', settings={ key: key }),
          $.interfaces.inspector.ip(type='unspecified', settings={ key: key }),
        ],
      },
      length: {
        // checks if data is equal to zero.
        //
        // use with the ANY / ALL operator to match empty data.
        // use with the NONE operator to match non-empty data.
        eq_zero(key=null):
          $.interfaces.inspector.length(type='equals', value=0, settings={ key: key }),
        // checks if data is greater than zero.
        //
        // use with the ANY / ALL operator to match non-empty data.
        // use with the NONE operator to match empty data.
        gt_zero(key=null):
          $.interfaces.inspector.length(type='greater_than', value=0, settings={ key: key }),
      },
      strings: {
        contains(expression, key=null):
          $.interfaces.inspector.strings(type='contains', expression=expression, settings={ key: key }),
        equals(expression, key=null):
          $.interfaces.inspector.strings(type='equals', expression=expression, settings={ key: key }),
        starts_with(expression, key=null):
          $.interfaces.inspector.strings(type='starts_with', expression=expression, settings={ key: key }),
        ends_with(expression, key=null):
          $.interfaces.inspector.strings(type='ends_with', expression=expression, settings={ key: key }),
      },
    },
    operator: {
      ip: {
        // returns true if the key is a valid IP address and is not private
        public(key=null): $.interfaces.operator.none(
          $.patterns.inspector.ip.private(key=key)
          + [
            // the none operator combined with negation returns true if the key is a valid IP
            $.interfaces.inspector.ip(type='valid', settings={ key: key, negate: true }),
          ]
        ),
        // returns true if the key is a private IP address
        private(key=null): $.interfaces.operator.any($.patterns.inspector.ip.private(key=key)),
      },
    },
    processor: {
      // replaces a condition in one or more processors.
      //
      // by default this will not replace a condition if the
      // processor(s) have no condition, but this can be overriden
      // by setting force to true.
      replace_condition(processor, condition, force=false): {
        local p = if !std.isArray(processor)
        then [processor]
        else processor,

        processor: [
          if force || std.objectHas(p.settings, 'condition')
          then std.mergePatch(p, { settings: { condition: condition } })
          else p

          for p in $.helpers.flatten_processors(p)
        ],
      },
      // executes one or more processors if key is not empty.
      //
      // if negate is set to true, then this executes the processor(s)
      // if key is empty.
      if_not_empty(processor, key, set_key=null, negate=false): {
        local i = if negate == false
        then $.patterns.inspector.length.gt_zero(key=key)
        else $.patterns.inspector.length.eq_zero(key=key),
        local c = $.interfaces.operator.all(i),

        processor: $.helpers.flatten_processors(
          $.patterns.processor.replace_condition(processor, c, force=true)
        ),
      },
      // performs a "move" by copying and deleting keys.
      move(key, set_key, condition=null): {
        processor: $.interfaces.processor.pipeline(
          // @this requires special handling because the delete processor cannot
          // delete complex objects.
          //
          // this works by copying the object into a metadata key, replacing the
          // object with empty data, then copying the metadata key into the
          // object.
          processors=if key == '@this'
          then [
            $.interfaces.processor.copy(settings={ set_key: '!metadata move' }),
            $.interfaces.processor.copy(settings={ key: '!metadata __null' }),
            $.interfaces.processor.copy(settings={ key: '!metadata move', set_key: set_key }),
          ]
          else [
            $.interfaces.processor.copy(settings={ key: key, set_key: set_key }),
            $.interfaces.processor.delete(settings={ key: key }),
          ],
          settings={ condition: condition },
        ),
      },
      copy: {
        // copies one or more keys into an array.
        //
        // apply a condition using the pipeline processor:
        //  local c = foo,
        //  local p = $.interfaces.processor.pipeline(processors=into_array(...).processors, condition=c),
        //  processor: $.interfaces.processor.apply(p)
        //
        // embed within other processor arrays by appending:
        //  processors: [
        //    ...,
        //    ...,
        // ] + into_array(...).processors
        into_array(keys, set_key, condition=null): {
          local opts = $.interfaces.processor.copy,

          processor: $.interfaces.processor.pipeline([
            $.interfaces.processor.copy(settings={ key: key, set_key: $.helpers.key.append_array(set_key) })
            for key in keys
          ], settings={ condition: condition }),
        },
      },
      dns: {
        // queries the Team Cymru Malware Hash Registry (https://www.team-cymru.com/mhr).
        //
        // MHR enriches hash data with a summary of results from anti-virus engines.
        // this patterns will cause significant latency in a data pipeline and should
        // be used in combination with a caching deployment patterns
        query_team_cymru_mhr(key, set_key='!metadata dns.query_team_cymru_mhr', condition=null): {
          local mhr_query = '!metadata query_team_cymru_mhr',
          local mhr_response = '!metadata response_team_cymru_mhr',

          processor: $.interfaces.processor.pipeline([
            // creates the MHR query domain by concatenating the key with the MHR service domain
            $.interfaces.processor.copy(settings={ key: key, set_key: $.helpers.key.append_array(mhr_query) }),
            $.interfaces.processor.insert(value='hash.cymru.com', settings={ set_key: $.helpers.key.append_array(mhr_query) }),
            $.interfaces.processor.join(separator='.', settings={ key: mhr_query, set_key: mhr_query }),
            // performs MHR query and parses returned value `["epoch" "hits"]` into object `{"team_cymru":{"epoch":"", "hits":""}}`
            $.interfaces.processor.dns(type='query_txt', settings={ key: mhr_query, set_key: mhr_response }),
            $.interfaces.processor.split(separator=' ', settings={ key: $.helpers.key.get_element(mhr_response, 0), set_key: mhr_response }),
            $.interfaces.processor.copy(settings={ key: $.helpers.key.get_element(mhr_response, 0), set_key: $.helpers.key.append(set_key, 'epoch') }),
            $.interfaces.processor.copy(settings={ key: $.helpers.key.get_element(mhr_response, 1), set_key: $.helpers.key.append(set_key, 'hits') }),
            // converts values from strings to integers
            $.interfaces.processor.convert(type='int', settings={
              key: $.helpers.key.append(set_key, 'epoch'),
              set_key: $.helpers.key.append(set_key, 'epoch'),
            }),
            $.interfaces.processor.convert(type='int', settings={ key: $.helpers.key.append(set_key, 'hits'), set_key: $.helpers.key.append(set_key, 'hits') }),
            // delete remaining keys
            $.interfaces.processor.delete(settings={ key: mhr_query }),
            $.interfaces.processor.delete(settings={ key: mhr_response }),
          ], settings={ condition: condition }),
        },
      },
      drop: {
        // randomly drops data.
        //
        // this can be used for integration testing when full load is not required.
        random: {
          local c = $.interfaces.operator.all($.interfaces.inspector.random),
          processor: $.interfaces.processor.drop(settings={ condition: c }),
        },
      },
      hash: {
        // hashes data using the SHA-256 algorithm.
        //
        // this patterns dynamically supports objects, plaintext data, and binary data.
        data(set_key='!metadata hash.data', algorithm='sha256'): {
          local hash = $.interfaces.processor.hash(algorithm=algorithm),

          // where data is temporarily stored during hashing
          local key = '!metadata data',

          local is_plaintext = $.interfaces.inspector.content(type='text/plain; charset=utf-8', settings={ key: key }),
          local is_json = $.interfaces.inspector.json_valid(),
          local not_json = $.interfaces.inspector.json_valid(settings={ negate: true }),

          processor: [
            // copies data to metadata for hashing
            $.interfaces.processor.copy(settings={ set_key: key }),
            // if data is an object, then hash its contents
            $.interfaces.processor.hash(algorithm=algorithm, settings={ key: '@this', set_key: set_key, condition: $.interfaces.operator.all([is_plaintext, is_json]) }),
            // if data is not an object but is plaintext, then hash it without decoding
            $.interfaces.processor.hash(algorithm=algorithm, settings={ key: key, set_key: set_key, condition: $.interfaces.operator.all([is_plaintext, not_json]) }),
            // if data is not plaintext, then decode and hash it
            $.interfaces.processor.pipeline([
              $.interfaces.processor.base64(direction='from'),
              $.interfaces.processor.hash(algorithm=algorithm),
            ], settings={ key: key, set_key: set_key, condition: $.interfaces.operator.none([is_plaintext]) }),
            // delete copied data
            $.interfaces.processor.delete(settings={ key: key }),
          ],
        },
      },
      ip_database: {
        // performs lookup for any public IP address in any IP enrichment database.
        lookup_address(key, set_key='!metadata ip_database.lookup_address', options=null): {
          assert options != null : 'ip_database.lookup_address options cannot be null',

          // only performs lookups against public IP addresses
          local c = $.patterns.operator.ip.public(key),

          processor: $.interfaces.processor.ip_database(options=options, settings={ key: key, set_key: set_key, condition: c }),
        },
      },
      time: {
        // generates current time.
        now(set_key='!metadata time.now', set_format=$.defaults.processor.time.set_format, condition=null): {
          processor: $.interfaces.processor.time(format='now', set_format=set_format, settings={ set_key: set_key, condition: condition }),
        },
      },
    },
  },
}
