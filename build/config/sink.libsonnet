{
  aws_dynamodb(table, key): {
    type: 'aws_dynamodb',
    settings: { table: table, key: key },
  },
  aws_kinesis(stream, partition='', partition_key='', shard_redistribution=false): {
    type: 'aws_kinesis',
    settings: { stream: stream, partition: partition, partition_key: partition_key, shard_redistribution: shard_redistribution },
  },
  aws_kinesis_firehose(stream): {
    type: 'aws_kinesis_firehose',
    settings: { stream: stream },
  },
  aws_s3(bucket, prefix='', prefix_key=''): {
    type: 'aws_s3',
    settings: { bucket: bucket, prefix: prefix, prefix_key: prefix_key },
  },
  aws_sqs(queue): {
    type: 'aws_sqs',
    settings: { queue: queue },
  },
  grpc(server, timeout='', certificate=''): {
    type: 'grpc',
    settings: { server: server, timeout: timeout, certificate: certificate },
  },
  http(url, headers=[], headers_key=''): {
    type: 'http',
    settings: { url: url, headers: headers, headers_key: headers_key },
  },
  stdout: {
    type: 'stdout',
  },
  sumologic(url, category='', category_key=''): {
    type: 'sumologic',
    settings: { url: url, category: category, category_key: category_key },
  },
}
