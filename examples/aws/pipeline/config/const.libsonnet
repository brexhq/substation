{
  // change uuid to match the resource created by Terraform
  s3_bucket: 'uuid-substation',
  // supports: json, text, data
  file_format: {
    type: 'json',
  },
  // supports: gzip, snappy, zstd
  file_compression: {
    type: 'gzip',
  },
}
