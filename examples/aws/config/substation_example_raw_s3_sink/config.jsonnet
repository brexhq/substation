local sink = import '../../../../build/config/sink.libsonnet';

{
  // writes objects to this S3 path: substation-example-raw/example/2022/01/01/*
  sink: sink.aws_s3(bucket='substation-example-raw', prefix='example'),
  transform: {
    type: 'transfer',
  },
}
