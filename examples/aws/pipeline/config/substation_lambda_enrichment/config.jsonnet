local sub = import '../../../../../build/config/substation.libsonnet';

local dns = import 'dns.libsonnet';

{
  // Substation microservices must use the gRPC localhost server.
  sink: sub.interfaces.sink.grpc(settings={server:'localhost:50051'}),
  // use the batch transform to modify data pushed to the processed Kinesis Data Stream.
  // processors are imported and compiled from local libsonnet files.
  transform: {
    type: 'batch',
    settings: {
      processors:
        dns.processors
    },
  },
}
