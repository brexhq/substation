local sinklib = import '../../build/config/sink.libsonnet';

{
  sink: sinklib.grpc(server='localhost:50051'),
  transform: {
    type: 'transfer'
  },
}
