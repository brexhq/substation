local sinklib = import '../../../../config/sink.libsonnet';

{
  sink: sinklib.kinesis(stream='substation_example_raw'),
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
