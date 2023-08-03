local sub = import '../../../../../build/config/substation.libsonnet';

local const = import '../const.libsonnet';

{
  transforms: [
    sub.interfaces.transform.send.aws_s3(
      settings={ 
        bucket: const.s3_bucket, 
        file_path: {
          prefix: 'raw',
          time_format: '2006/01/02',
          uuid: true,
          extension: true,
        },
        file_format: const.file_format,
        file_compression: const.file_compression,
      }
    )
  ]
}
