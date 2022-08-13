local event = import './event.libsonnet';
local sinklib = import '../../config/sink.libsonnet';

{
  sink: sinklib.stdout,
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors
    },
  },
}
