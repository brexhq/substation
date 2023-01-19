local sub = import '../../../../../build/config/substation.libsonnet';

local event = import 'event.libsonnet';
local enrich = import 'enrich.libsonnet';

{
  sink: sub.interfaces.sink.aws_kinesis(settings={stream:'substation_processed'}),
  // use the batch transform to modify data pushed to the processed Kinesis Data Stream.
  // processors are imported and compiled from local libsonnet files.
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors
        + enrich.processors,
      // + foo.processors
      // + bar.processors
      // + baz.processors
    },
  },
}
