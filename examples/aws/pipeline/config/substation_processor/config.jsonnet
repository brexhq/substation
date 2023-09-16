local sub = import '../../../../../build/config/substation.libsonnet';

local enrich = import 'enrich.libsonnet';

{
  transforms: [
    sub.transform.hash.sha256(
      settings={ object: {key: '@this', set_key: 'event.hash' } }
    ),
    sub.transform.time.now(
      settings={ set_key: 'event.created' }
    ),

    sub.transform.send.aws.kinesis_data_stream(
      settings={stream:'substation_processed'},
    )
  ]
}
