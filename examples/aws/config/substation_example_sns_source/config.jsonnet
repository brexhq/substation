local sink = import '../../../../build/config/sink.libsonnet';

{
  sink: sink.aws_kinesis(stream='substation_example_raw'),
  transform: {
    type: 'transfer',
  },
}
