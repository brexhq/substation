local sub = import '../../../../../build/config/substation.libsonnet';

{
  sink: sub.interfaces.sink.stdout,
  transform: {
    type: 'noop',
  },
}
