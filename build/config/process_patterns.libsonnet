local inspector = import 'inspector.libsonnet';
local inspectorPatterns = import 'inspector_patterns.libsonnet';

local operatorPatterns = import 'operator_patterns.libsonnet';

local process = import 'process.libsonnet';

{
  dns: {
    // queries the Team Cymru Malware Hash Registry (https://www.team-cymru.com/mhr).
    //
    // MHR enriches hash data with a summary of results from anti-virus engines.
    // this pattern will cause significant latency in a data pipeline and should
    // be used in combination with a caching deployment pattern
    query_team_cymru_mhr(key, set_key='!metadata team_cymru_mhr'): {
      processors: [
        // creates the MHR query domain by concatenating the key with the MHR service domain
        process.apply(options=process.copy,
                      key=key,
                      set_key='!metadata query_team_cymru_mhr.-1'),
        process.apply(options=process.insert(value='hash.cymru.com'),
                      set_key='!metadata query_team_cymru_mhr.-1'),
        process.apply(options=process.join(separator='.'),
                      key='!metadata query_team_cymru_mhr',
                      set_key='!metadata query_team_cymru_mhr'),
        // performs MHR query and parses returned value `["epoch" "hits"]` into object `{"team_cymru":{"epoch":"", "hits":""}}`
        process.apply(options=process.dns(type='query_txt'),
                      key='!metadata query_team_cymru_mhr',
                      set_key='!metadata response_team_cymru_mhr'),
        process.apply(options=process.split(separator=' '),
                      key='!metadata response_team_cymru_mhr.0',
                      set_key='!metadata response_team_cymru_mhr'),
        process.apply(options=process.copy,
                      key='!metadata response_team_cymru_mhr.0',
                      set_key=set_key + '.epoch'),
        process.apply(options=process.copy,
                      key='!metadata response_team_cymru_mhr.1',
                      set_key=set_key + '.hits'),
        // converts values from strings to integers
        process.apply(options=process.convert(type='int'),
                      key=set_key + '.epoch',
                      set_key=set_key + '.epoch'),
        process.apply(options=process.convert(type='int'),
                      key=set_key + '.hits',
                      set_key=set_key + '.hits'),
        // // delete remaining keys
        process.apply(options=process.delete,
                      key='!metadata query_team_cymru_mhr'),
        process.apply(options=process.delete,
                      key='!metadata response_team_cymru_mhr'),
      ],
    },
  },
  drop: {
    // randomly drops data.
    //
    // this can be used for integration testing when full load is not required.
    random: {
      local op = operatorPatterns.all([inspector.inspect(inspector.random)]),
      processor: process.apply(options=process.drop, condition=op),
    },
  },
  copy: {
    into_array(keys, set_key): {
      processors: [
        process.apply(process.copy, key=key, set_key=set_key + '.-1')
        for key in keys
      ],
    },
  },
  hash: {
    // hashes data using the SHA-256 algorithm.
    //
    // this pattern dynamically supports objects, plaintext data, and binary data.
    data(set_key='!metadata hash', algorithm='sha256'): {
      local hash_opts = process.hash(algorithm=algorithm),

      // data is temporarily stored during hashing
      local key = '!metadata data',

      local is_plaintext = inspector.inspect(inspector.content(type='text/plain; charset=utf-8'), key=key),
      local is_json = inspector.inspect(inspector.json_valid),
      local is_not_json = inspector.inspect(inspector.json_valid, negate=true),

      processors: [
        // copies data to metadata for hashing
        process.apply(options=process.copy, set_key=key),
        // if data is an object, then hash the object's content
        process.apply(options=hash_opts,
                      key='@this',
                      set_key=set_key,
                      condition=operatorPatterns.all([is_plaintext, is_json])),
        // if data is not an object but is plaintext, then hash the data as-is
        process.apply(options=hash_opts,
                      key=key,
                      set_key=set_key,
                      condition=operatorPatterns.all([is_plaintext, is_not_json])),
        // if data is not plaintext, then decode and hash the data
        process.apply(
          options=process.pipeline([
            process.apply(options=process.base64(direction='from')),
            process.apply(options=hash_opts),
          ]),
          key=key,
          set_key=set_key,
          condition=operatorPatterns.none([is_plaintext])
        ),
        // delete copied data
        process.apply(options=process.delete, key=key),
      ],
    },
  },
  ip_database: {
    // performs lookup for any public IP address in any IP enrichment database.
    lookup_public_address(key, set_key, options): {
      local ip_database_opts = process.ip_database(options),
      local op = operatorPatterns.none(inspectorPatterns.ip.private(key)),

      processor: process.apply(options=ip_database_opts, key=key, set_key=set_key, condition=op),
    },
  },
}
