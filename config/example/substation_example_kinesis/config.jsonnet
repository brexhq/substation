local sinklib = import '../../sink.libsonnet';
local event = import './event.libsonnet';

{
  sink: sinklib.kinesis(stream='substation_processed_example'),
  // use the process transform to modify data pushed to the processed Kinesis Data Stream; processors are imported and compiled from local libsonnet files
  transform: {
    type: 'process',
    settings: {
      processors:
        event.processors
        // + foo.processors
        // + bar.processors
        // + baz.processors
    },
  },
}
