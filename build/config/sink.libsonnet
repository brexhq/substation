{
  dynamodb(table, items_key): {
    type: 'dynamodb',
    settings: { table: table, items_key: items_key },
  },
  http(url, headers=[], headers_key=''): {
    type: 'http',
    settings: { url: url, headers: headers, headers_key: headers_key },
  },
  kinesis(stream, partition='', partition_key='', shard_redistribution=false): {
    type: 'kinesis',
    settings: { stream: stream, partition: partition, partition_key: partition_key, shard_redistribution: shard_redistribution },
  },
  firehose(stream): {
    type: 'firehose',
    settings: { stream: stream },
  },
  grpc(server, timeout='', certificate=''): {
    type: 'grpc',
    settings: { server: server, timeout: timeout, certificate: certificate },
  },
  s3(bucket, prefix='', prefix_key=''): {
    type: 's3',
    settings: { bucket: bucket, prefix: prefix, prefix_key: prefix_key },
  },
  stdout: {
    type: 'stdout',
  },
  sumologic(url, category='', category_key=''): {
    type: 'sumologic',
    settings: { url: url, category: category, category_key: category_key },
  },
  sqs(queue): {
    type: 'sqs',
    settings: { queue: queue },
  },
}
