local sink = import '../../build/config/sink.libsonnet';

{
  sink: sink.grpc(server='localhost:50051'),
  transform: {
    type: 'transfer',
  },
}
