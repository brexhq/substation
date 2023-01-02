local helpers = import 'helpers.libsonnet';
local lib = import 'interfaces.libsonnet';

{
  inspect: {
    // negates any inspector
    negate(inspector): std.mergePatch(inspector, { settings: { negate: true } }),
    ip: {
      // checks if an IP address is private.
      //
      // use with the ANY operator to match private IP addresses.
      // use with the NONE operator to match public IP addresses.
      private(key=null): [
        lib.inspector.inspect(lib.inspector.ip(type='loopback'), key=key),
        lib.inspector.inspect(lib.inspector.ip(type='multicast'), key=key),
        lib.inspector.inspect(lib.inspector.ip(type='multicast_link_local'), key=key),
        lib.inspector.inspect(lib.inspector.ip(type='private'), key=key),
        lib.inspector.inspect(lib.inspector.ip(type='unicast_link_local'), key=key),
        lib.inspector.inspect(lib.inspector.ip(type='unspecified'), key=key),
      ],
    },
    length: {
      // checks if data is equal to zero.
      //
      // use with the ANY / ALL operator to match empty data.
      // use with the NONE operator to match non-empty data.
      eq_zero(key=null): lib.inspector.inspect(
        lib.inspector.length(type='equals', value=0), key=key,
      ),
      // checks if data is greater than zero.
      //
      // use with the ANY / ALL operator to match non-empty data.
      // use with the NONE operator to match empty data.
      gt_zero(key=null): lib.inspector.inspect(
        lib.inspector.length(type='greater_than', value=0), key=key,
      ),
    },
    strings: {
      contains(expression, key=null): lib.inspector.inspect(
        lib.inspector.strings(type='contains', expression=expression), key=key,
      ),
      equals(expression, key=null): lib.inspector.inspect(
        lib.inspector.strings(type='equals', expression=expression), key=key,
      ),
      starts_with(expression, key=null): lib.inspector.inspect(
        lib.inspector.strings(type='starts_with', expression=expression), key=key,
      ),
      ends_with(expression, key=null): lib.inspector.inspect(
        lib.inspector.strings(type='ends_with', expression=expression), key=key,
      ),
    },
  },
  operate: {
    ip: {
      // returns true if the key is a valid IP address and is not private
      public(key=null): lib.operator.none(
        $.inspect.ip.private(key=key)
        + [
          // the none operator combined with negation returns true if the key is a valid IP
          lib.inspector.inspect(lib.inspector.ip(type='valid'), key=key, negate=true),
        ]
      ),
      // returns true if the key is a private IP address
      private(key=null): lib.operator.any($.inspect.ip.private(key=key)),
    },
  },
  process: {
    // replaces a condition in one or more processors.
    //
    // by default this will not replace a condition if the
    // processor(s) have no condition, but this can be overriden
    // by setting force to true.
    replace_condition(processors, condition, force=false): {
      local p = if !std.isArray(processors)
      then [processors]
      else processors,

      processors: [
        if force || p.settings.condition != null
        then std.mergePatch(p, { settings: { condition: condition } })
        else p

        for p in helpers.flatten_processors(p)
      ],
    },
    // executes one or more processors if key is not empty.
    //
    // if negate is set to true, then this executes the processor(s)
    // if key is empty.
    if_not_empty(processors, key, set_key=null, negate=false): {
      local i = if negate == false
      then $.inspect.length.gt_zero(key=key)
      else $.inspect.length.eq_zero(key=key),
      local c = lib.operator.all(i),

      processors: helpers.flatten_processors(
        $.process.replace_condition(processors, c, force=true)
      ),
    },
    // performs a "move" by copying and deleting keys.
    move(key, set_key, condition=null): {
      processor: lib.process.apply(
        lib.process.pipeline(
          // @this requires special handling because the delete processor cannot
          // delete complex objects.
          //
          // this works by copying the object into a metadata key, replacing the
          // object with empty data, then copying the metadata key into the
          // object.
          if key == '@this'
          then [
            lib.process.apply(lib.process.copy, set_key='!metadata move'),
            lib.process.apply(lib.process.copy, set_key='!metadata move'),
            lib.process.apply(lib.process.copy, key='!metadata move.null'),
            lib.process.apply(lib.process.copy, key='!metadata move', set_key=set_key),
          ]
          else [
            lib.process.apply(lib.process.copy, key=key, set_key=set_key),
            lib.process.apply(lib.process.delete, key=key),
          ]
        ),
        condition=condition,
      ),
    },
    copy: {
      // copies one or more keys into an array.
      //
      // apply a condition using the pipeline processor:
      //  local c = foo,
      //  local p = lib.process.pipeline(processors=into_array(...).processors, condition=c),
      //  processor: lib.process.apply(p)
      //
      // embed within other processor arrays by appending:
      //  processors: [
      //    ...,
      //    ...,
      // ] + into_array(...).processors
      into_array(keys, set_key, condition=null): {
        local opts = lib.process.copy,

        processor: lib.process.apply(
          lib.process.pipeline([
            lib.process.apply(options=opts, key=key, set_key=helpers.key.append_array(set_key))
            for key in keys
          ]),
          condition=condition,
        ),
      },
    },
    dns: {
      // queries the Team Cymru Malware Hash Registry (https://www.team-cymru.com/mhr).
      //
      // MHR enriches hash data with a summary of results from anti-virus engines.
      // this pattern will cause significant latency in a data pipeline and should
      // be used in combination with a caching deployment pattern
      query_team_cymru_mhr(key, set_key='!metadata dns.query_team_cymru_mhr', condition=null): {
        local mhr_query = '!metadata query_team_cymru_mhr',
        local mhr_response = '!metadata response_team_cymru_mhr',

        processor: lib.process.apply(
          lib.process.pipeline([
            // creates the MHR query domain by concatenating the key with the MHR service domain
            lib.process.apply(
              options=lib.process.copy,
              key=key,
              set_key=helpers.key.append_array(mhr_query),
            ),
            lib.process.apply(
              options=lib.process.insert(value='hash.cymru.com'),
              set_key=helpers.key.append_array(mhr_query),
            ),
            lib.process.apply(
              options=lib.process.join(separator='.'),
              key=mhr_query,
              set_key=mhr_query,
            ),
            // performs MHR query and parses returned value `["epoch" "hits"]` into object `{"team_cymru":{"epoch":"", "hits":""}}`
            lib.process.apply(
              options=lib.process.dns(type='query_txt'),
              key=mhr_query,
              set_key=mhr_response,
            ),
            lib.process.apply(
              options=lib.process.split(separator=' '),
              key=helpers.key.get_element(mhr_response, 0),
              set_key=mhr_response,
            ),
            lib.process.apply(
              options=lib.process.copy,
              key=helpers.key.get_element(mhr_response, 0),
              set_key=helpers.key.append(set_key, 'epoch'),
            ),
            lib.process.apply(
              options=lib.process.copy,
              key=helpers.key.get_element(mhr_response, 1),
              set_key=helpers.key.append(set_key, 'hits')
            ),
            // converts values from strings to integers
            lib.process.apply(
              options=lib.process.convert(type='int'),
              key=helpers.key.append(set_key, 'epoch'),
              set_key=helpers.key.append(set_key, 'epoch')
            ),
            lib.process.apply(
              options=lib.process.convert(type='int'),
              key=helpers.key.append(set_key, 'hits'),
              set_key=helpers.key.append(set_key, 'hits')
            ),
            // delete remaining keys
            lib.process.apply(options=lib.process.delete, key=mhr_query),
            lib.process.apply(options=lib.process.delete, key=mhr_response),
          ]),
          condition=condition,
        ),
      },
    },
    drop: {
      // randomly drops data.
      //
      // this can be used for integration testing when full load is not required.
      random: {
        local c = lib.operator.all(lib.inspector.inspect(lib.inspector.random)),
        processor: lib.process.apply(options=lib.process.drop, condition=c),
      },
    },
    hash: {
      // hashes data using the SHA-256 algorithm.
      //
      // this pattern dynamically supports objects, plaintext data, and binary data.
      data(set_key='!metadata hash.data', algorithm='sha256'): {
        local hash_opts = lib.process.hash(algorithm=algorithm),

        // where data is temporarily stored during hashing
        local key = '!metadata data',

        local is_plaintext = lib.inspector.inspect(lib.inspector.content(type='text/plain; charset=utf-8'), key=key),
        local is_json = lib.inspector.inspect(lib.inspector.json_valid),
        local is_not_json = lib.inspector.inspect(lib.inspector.json_valid, negate=true),

        processors: [
          // copies data to metadata for hashing
          lib.process.apply(options=lib.process.copy, set_key=key),
          // if data is an object, then hash its contents
          lib.process.apply(options=hash_opts,
                            key='@this',
                            set_key=set_key,
                            condition=lib.operator.all([is_plaintext, is_json])),
          // if data is not an object but is plaintext, then hash it without decoding
          lib.process.apply(options=hash_opts,
                            key=key,
                            set_key=set_key,
                            condition=lib.operator.all([is_plaintext, is_not_json])),
          // if data is not plaintext, then decode and hash it
          lib.process.apply(
            options=lib.process.pipeline([
              lib.process.apply(options=lib.process.base64(direction='from')),
              lib.process.apply(options=hash_opts),
            ]),
            key=key,
            set_key=set_key,
            condition=lib.operator.none([is_plaintext])
          ),
          // delete copied data
          lib.process.apply(options=lib.process.delete, key=key),
        ],
      },
    },
    ip_database: {
      // performs lookup for any public IP address in any IP enrichment database.
      lookup_address(key, set_key='!metadata ip_database.lookup_address', options=null): {
        assert options != null : 'options cannot be null',

        local opts = lib.process.ip_database(options),
        // only performs lookups against public IP addresses
        local op = lib.operator.ip.public(key),

        processor: lib.process.apply(options=opts, key=key, set_key=set_key, condition=op),
      },
    },
    time: {
      // generates current time.
      now(set_key='!metadata time.now', set_format=null, condition=null): {
        local opts = lib.process.time(format='now', set_format=set_format),
        processor: lib.process.apply(options=opts, set_key=set_key, condition=condition),
      },
    },
  },
}
