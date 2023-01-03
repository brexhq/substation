local sub = import '../../../../build/config/substation.libsonnet';

{
  sink: sub.interfaces.sink.aws_kinesis(settings={stream:'substation_example_raw'}),
  // use the transfer transform so data is not modified in transit
  transform: {
    type: 'transfer',
  },
}
