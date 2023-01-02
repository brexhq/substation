local lib = import '../../../../build/config/interfaces.libsonnet';

{
  sink: lib.sink.aws_kinesis(stream='substation_example_raw'),
  // use the transfer transform so data is not modified in transit
  transform: {
    type: 'transfer',
  },
}
