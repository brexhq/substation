local sinklib = import '../../../../build/config/sink.libsonnet';

{
  sink: sinklib.kinesis(stream='substation_example_raw'),
  transform: {
    type: 'transfer',
  },
}