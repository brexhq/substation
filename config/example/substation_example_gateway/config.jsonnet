local sinklib = import '../../sink.libsonnet';

{
  sink: sinklib.kinesis(stream='substation_raw_example'),
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
