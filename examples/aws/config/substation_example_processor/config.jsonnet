local sinklib = import '../../../../build/config/sink.libsonnet';

local event = import './event.libsonnet';

{
  sink: sinklib.kinesis(stream='substation_example_processed'),
  // use the process transform to modify data pushed to the processed Kinesis Data Stream; processors are imported and compiled from local libsonnet files
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors,
      // + foo.processors
      // + bar.processors
      // + baz.processors
    },
  },
}
