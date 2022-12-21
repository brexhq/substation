local condition = import 'condition.libsonnet';
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
    query_team_cymru_mhr(key, set_key='!metadata team_cymru_mhr'): [{
      processors: [
        // creates the MHR query domain by concatenating the key with the MHR service domain
        process.process(options=process.copy, key=key, set_key='!metadata query_team_cymru_mhr.-1'),
        process.process(options=process.insert(value='hash.cymru.com'), set_key='!metadata query_team_cymru_mhr.-1'),
        process.process(options=process.join(separator='.'), key='!metadata query_team_cymru_mhr', set_key='!metadata query_team_cymru_mhr'),
        // performs MHR query and parses returned value `["epoch" "hits"]` into object `{"team_cymru":{"epoch":"", "hits":""}}`
        process.process(options=process.dns(type='query_txt'), key='!metadata query_team_cymru_mhr', set_key='!metadata response_team_cymru_mhr'),
        process.process(options=process.split(separator=' '), key='!metadata response_team_cymru_mhr.0', set_key='!metadata response_team_cymru_mhr'),
        process.process(options=process.copy, key='!metadata response_team_cymru_mhr.0', set_key=set_key + '.epoch'),
        process.process(options=process.copy, key='!metadata response_team_cymru_mhr.1', set_key=set_key + '.hits'),
        // converts values from strings to integers
        process.process(options=process.convert(type='int'), key=set_key + '.epoch', set_key=set_key + '.epoch'),
        process.process(options=process.convert(type='int'), key=set_key + '.hits', set_key=set_key + '.hits'),
        // // delete remaining keys
        process.process(options=process.delete, key='!metadata query_team_cymru_mhr'),
        process.process(options=process.delete, key='!metadata response_team_cymru_mhr'),
      ],
    }],
  },
  drop: {
    // randomly drops data.
    //
    // this can be used for integration testing when full load is not required.
    random: [
      {
        local op = operatorPatterns.and([condition.inspector.random]),
        processors: [
          process.process(options=process.drop, condition=op),
        ],
      },
    ],
  },
  hash: {
    // hashes data using the SHA-256 algorithm.
    //
    // this pattern dynamically supports plaintext and binary data.
    data(set_key='!metadata hash', algorithm='sha256'): [
      {
        local hash = process.hash(algorithm=algorithm),

        // data is temporarily stored during hashing
        local key = '!metadata data',

        // plaintext content match determines how data should be hashed
        local is_plaintext = condition.inspector(
          condition.content(type='text/plain; charset=utf-8'), key=key
        ),

        // if data is not plaintext, then it is treated as binary data
        local pipeline = process.pipeline([
          process.process(options=process.base64(direction='from')),
          process.process(options=hash),
        ]),

        processors: [
          // copies data to metadata for hashing
          process.process(options=process.copy, set_key=key),
          // applies plaintext hashing
          process.process(options=hash, condition=operatorPatterns.and([is_plaintext]), key='@this', set_key=set_key),
          // applies non-plaintext (binary) hashing
          process.process(options=pipeline, condition=operatorPatterns.nand([is_plaintext]), key=key, set_key=set_key),
          // delete remaining data
          process.process(options=process.delete, key=key),
        ],
      },
    ],
  },
  ip_database: {
    // performs lookup for any public IP address in any IP enrichment database.
    lookup_public_address(key, set_key, ipdb_options): [
      {
        local ipdb_opts = process.ip_database(ipdb_options),
        local op = operatorPatterns.nand(inspectorPatterns.ip.private(key)),

        processors: [
          process.process(options=ipdb_opts, condition=op, key=key, set_key=set_key),
        ],
      },
    ],
  },
}
