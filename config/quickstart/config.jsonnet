local event = import './event.libsonnet';
local sinklib = import '../sink.libsonnet';

{
  sink: sinklib.stdout,
  transform: {
    type: 'process',
    settings: {
      processors:
        event.processors
    },
  },
}
