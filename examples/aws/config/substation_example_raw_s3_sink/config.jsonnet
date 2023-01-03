local sub = import '../../../../build/config/substation.libsonnet';

{
  // writes objects to this S3 path: substation-example-raw/example/2022/01/01/*
  sink: sub.interfaces.sink.aws_s3(bucket='substation-example-raw', prefix='example'),
  // use the transfer transform so data is not modified in transit
  transform: {
    type: 'transfer',
  },
}
