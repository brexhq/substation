local sinklib = import '../../sink.libsonnet';

{
  sink: sinklib.kinesis(stream='substation_example_raw'),
  transform: {
    type: 'transfer',
  },
}
