local sub = import '../../../../../build/config/substation.libsonnet';

local const = import '../const.libsonnet';

{
  sink: sub.interfaces.sink.aws_s3(
    settings={
      // change S3 bucket uuid in const to match the resource created by Terraform
      bucket: const.s3_bucket,
      // file path becomes processed/2006/01/02/uuid.extension
      file_path: {
        prefix: 'processed',
        time_format: '2006/01/02',
        uuid: true,
        extension: true,
      },
      file_format: const.file_format,
      file_compression: const.file_compression,
    }
  ),
  // use the transfer transform so data is not modified in transit
  transform: {
    type: 'transfer',
  },
}
