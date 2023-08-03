{
  // Change uuid to match the resource created by Terraform.
  s3_bucket: 'uuid-substation',
  // Supports: json, text, data.
  file_format: {
    type: 'json',
  },
  // Supports: gzip, snappy, zstd.
  file_compression: {
    type: 'gzip',
  },
}
