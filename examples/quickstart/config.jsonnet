local sink = import '../../build/config/sink.libsonnet';

local event = import './event.libsonnet';

{
  sink: sink.stdout,
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors,
    },
  },
}
