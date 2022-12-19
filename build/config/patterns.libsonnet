// functions in this file contain pre-configured conditions and processors that represent commonly used patterns across many data pipelines.

local conditionlib = import './condition.libsonnet';
local ipdatabaselib = import './ip_database.libsonnet';
local processlib = import './process.libsonnet';

{
  dns: {
    // writes results from the Team Cymru Malware Hash Registry to '!metadata team_cymru_mhr'
    // this pattern will cause significant data latency in a data pipeline and should be used in combination with a caching deployment pattern
    // https://www.team-cymru.com/mhr
    query_team_cymru_mhr(input): [{
      processors: [
        // creates the MHR query domain by concatenating the input with the MHR service domain
        processlib.copy(input=input, output='!metadata query_team_cymru_mhr.-1'),
        processlib.insert(output='!metadata query_team_cymru_mhr.-1', value='hash.cymru.com'),
        processlib.concat(input='!metadata query_team_cymru_mhr', output='!metadata query_team_cymru_mhr', separator='.'),
        // performs MHR query and parses returned value `["epoch" "hits"]` into JSON `{"team_cymru":{"epoch":"", "hits":""}}` 
        processlib.dns(input='!metadata query_team_cymru_mhr', output='!metadata response_team_cymru_mhr', _function='query_txt'),
        processlib.split(input='!metadata response_team_cymru_mhr.0', output='!metadata response_team_cymru_mhr', separator=' '),
        processlib.copy(input='!metadata response_team_cymru_mhr.0', output='!metadata team_cymru_mhr.epoch'),
        processlib.copy(input='!metadata response_team_cymru_mhr.1', output='!metadata team_cymru_mhr.hits'),
        // converts JSON values from strings into integers
        processlib.convert(input='!metadata team_cymru_mhr.epoch', output='!metadata team_cymru_mhr.epoch', type='int'),
        processlib.convert(input='!metadata team_cymru_mhr.hits', output='!metadata team_cymru_mhr.hits', type='int'),
        // delete remaining keys
        processlib.delete(input='!metadata query_team_cymru_mhr'),
        processlib.delete(input='!metadata response_team_cymru_mhr'),
      ],
    }],
  },
  drop: {
    // drops randomly selected data. this can be useful for integration tests.
    random_data: [{
      local conditions = [
        conditionlib.random,
      ],
      processors: [
        processlib.drop(condition_operator='or', condition_inspectors=conditions),
      ],
    }],
  },
  hash: {
    // hashes data with the SHA-256 function and stores the hash in metadata
    data: [
      {
        local is_plaintext = conditionlib.content(type='text/plain; charset=utf-8'),
        processors: [
          // copies data to metadata
          processlib.copy(
            input='',
            output='!metadata data'
          ),

          // if data is plaintext, then directly hash it
          processlib.hash(
            input='!metadata data',
            output='!metadata hash',
            algorithm='sha256',
            condition_operator='or',
            condition_inspectors=[is_plaintext]
          ),

          // if data is not plaintext, then hash it via a pipeline. binary data stored in JSON is encoded as base64, so it is first decoded then hashed.
          processlib.pipeline(
            input='!metadata data',
            output='!metadata hash',
            processors=[
              processlib.base64('', '', direction='from'),
              processlib.hash('', '', algorithm='sha256'),
            ],
            condition_operator='nor',
            condition_inspectors=[is_plaintext]
          ),

          // delete the data stored in metadata
          processlib.delete(
            input='!metadata data'
          ),
        ],
      },
    ],
  },
  ip_database: {
    // performs lookup for any valid, public IP address in any IP enrichment database
    lookup_public_address(input, output, db_options): [{
      local conditions = [
        conditionlib.ip.valid(input),
        conditionlib.ip.loopback(input, negate=true),
        conditionlib.ip.multicast(input, negate=true),
        conditionlib.ip.multicast_link_local(input, negate=true),
        conditionlib.ip.private(input, negate=true),
        conditionlib.ip.unicast_link_local(input, negate=true),
        conditionlib.ip.unspecified(input, negate=true),
      ],
      processors: [
        processlib.ip_database(input=input, output=output, database_options=db_options, condition_operator='and', condition_inspectors=conditions),
      ],
    }],
  },
}
