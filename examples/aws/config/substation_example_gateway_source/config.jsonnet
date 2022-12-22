local sink = import '../../../../build/config/sink.libsonnet';

{
  sink: sink.kinesis(stream='substation_example_raw'),
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
