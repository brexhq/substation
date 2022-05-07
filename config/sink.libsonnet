{
  dynamodb(table, attributes, error_on_failure=false): {
    type: 'dynamodb',
    settings: {
      table: table,
      attributes: attributes,
      error_on_failure: error_on_failure,
    },
  },
  http(url, headers=[]): {
    type: 'http',
    settings: {
      url: url,
      headers: headers,
    },
  },
  kinesis(stream, partition_key=''): {
    type: 'kinesis',
    settings: {
      stream: stream,
      partition_key: partition_key,
    },
  },
  s3(bucket, prefix=''): {
    type: 's3',
    settings: {
      bucket: bucket,
      prefix: prefix,
    },
  },
  stdout: {
    type: 'stdout',
  },
  sumologic(url, category_key='', error_on_failure=false): {
    type: 's3',
    settings: {
      url: url,
      category_key: category_key,
      error_on_failure: error_on_failure,
    },
  },
}
