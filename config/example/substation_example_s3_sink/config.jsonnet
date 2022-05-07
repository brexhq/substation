local sinklib = import '../../sink.libsonnet';

{
  // writes objects to this S3 path: substation-example-sink/example/2022/01/01/*
  sink: sinklib.s3(bucket='substation-example-sink', prefix='example'),
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
