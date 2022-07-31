local sinklib = import '../../../config/sink.libsonnet';

{
  // writes objects to this S3 path: substation-example-processed/example/2022/01/01/*
  sink: sinklib.s3(bucket='substation-example-processed', prefix='example'),
  transform: {
    type: 'transfer',
  },
}
