local event = import './event.libsonnet';

{
  sink: {
    type: 'stdout',
  },
  transform: {
    type: 'process',
    settings: {
      processors:
        event.processors
    },
  },
}
