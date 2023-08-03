local sub = import '../../../../../build/config/substation.libsonnet';

local enrich = import 'enrich.libsonnet';

{
  transforms: [
    sub.interfaces.transform.proc.hash(
      settings={ key: '@this', set_key: 'event.hash' }
    ),
    sub.interfaces.transform.proc.time(
      settings={ set_key: 'event.created', format: 'now' }
    ),

    sub.interfaces.transform.send.aws_kinesis(
      settings={stream:'substation_processed'},
    )
  ]
}
