local sub = std.extVar('sub');

{
  transforms: [
    sub.tf.send.gcp.storage({
      // Write the message stream to the GCS bucket `substation-bucket`.
      gcp: { resource: 'projects/_/buckets/substation-bucket' },
      // Bucket objects are organized by time to the nearest hour and have a UUID filename.
      file_path: { time_format: '2006/01/02/15', uuid: true, suffix: '.jsonl.gz' },
      // This example formats the data as JSON Lines and compresses it with Gzip.
      aux_tforms: [
        sub.tf.agg.to.string({ separator: '\n' }),
        sub.tf.str.append({ suffix: '\n' }),
        sub.tf.fmt.to.gzip(),
      ],
    }),
  ],
}
