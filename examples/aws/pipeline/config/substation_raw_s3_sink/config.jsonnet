local sub = import '../../../../../build/config/substation.libsonnet';

local const = import '../const.libsonnet';

{
  // writes objects to this S3 path: uuid-substation/raw/2022/01/01/*
  // change uuid to match the resource created by Terraform
  sink: sub.interfaces.sink.aws_s3(settings={bucket:const.s3_bucket, prefix:'raw'}),
  // use the transfer transform so data is not modified in transit
  transform: {
    type: 'transfer',
  },
}
