// This example configures a storage class for the AWS S3 destination transform.
// The Glacier Instant Retrieval class is recommended for archival data that is
// compatible with Substation's serverless architecture; this class can be read
// directly by a Lambda function triggered by an SNS notification.
local sub = import '../../../../substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.aws.s3({
      // Glacier Instant Retrieval charges a minimum of 128KB per object, otherwise
      // the other values are set to impossibly high values to ensure all events are
      // written to the same file.
      batch: { size: 128 * 1000, count: 1000 * 1000, duration: '60m' },
      bucket_name: 'substation',
      storage_class: 'GLACIER_IR',  // Glacier Instant Retrieval.
      // S3 objects are organized by time to the nearest hour and have a UUID filename.
      file_path: { time_format: '2006/01/02/15', uuid: true, suffix: '.jsonl.gz' },
      // This example formats the data as JSON Lines and compresses it with Gzip.
      aux_tforms: [
        sub.tf.agg.to.string({ separator: '\n' }),
        sub.tf.str.append({ suffix: '\n' }),
        sub.tf.fmt.to.gzip()
      ],
    }),
  ],
}
