local sinklib = import '../../build/config/sink.libsonnet';
local event = import './event.libsonnet';

{
  sink: sinklib.stdout,
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors,
    },
  },
}
