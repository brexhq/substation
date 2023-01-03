local sub = import '../../build/config/substation.libsonnet';

{
  sink: sub.interfaces.sink.grpc(settings={server:'localhost:50051'}),
  transform: {
    type: 'transfer',
  },
}
