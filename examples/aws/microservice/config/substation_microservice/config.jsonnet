local sub = import '../../../../../build/config/substation.libsonnet';

local dns = import 'dns.libsonnet';

{
  // the sink must always be a gRPC server hosted on localhost:50051
  sink: sub.interfaces.sink.grpc(
    settings={server:'localhost:50051'}
  ),
  // use the batch transform to modify data sent to the service.
  transform: {
    type: 'batch',
    settings: {
      processors:
        dns.processors,
    },
  },
}
