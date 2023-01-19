local sub = import '../../../../../build/config/substation.libsonnet';

local const = import '../const.libsonnet';

{
  // writes objects to this S3 path: {const.s3_bucket}/processed/2022/01/01/*
  sink: sub.interfaces.sink.aws_s3(settings={bucket:const.s3_bucket, prefix:'processed'}),
  // use the transfer transform so data is not modified in transit
  transform: {
    type: 'transfer',
  },
}
