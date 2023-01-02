local lib = import '../../build/config/interfaces.libsonnet';

local event = import 'event.libsonnet';

{
  sink: lib.sink.stdout,
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors,
    },
  },
}
