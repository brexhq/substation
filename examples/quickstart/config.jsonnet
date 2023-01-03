local sub = import '../../build/config/substation.libsonnet';

local event = import 'event.libsonnet';

{
  sink: sub.interfaces.sink.stdout,
  transform: {
    type: 'batch',
    settings: {
      processors:
        event.processors,
    },
  },
}
