local sub = import '../../../../../../../../build/config/substation.libsonnet';

// This is a placeholder that must be replaced with the bucket produced by Terraform.
local bucket = 'c034c726-70bf-c397-81bd-c9a0d9e82371-substation';
local const = import 'const.libsonnet';

{
  concurrency: 1,
  transforms: [
    // XDR threat signals rely on the meta_switch transform to conditionally determine
    // if an event matches risk criteria. If the event matches, a threat signal is created.
    // This transform supports any combination of if-elif-else logic.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.any([
          sub.cnd.str.has({ obj: { src: 'user_name' }, value: 'root' }),
        ]),
        transform: sub.tf.meta.pipe({ transforms: const.threat_signal({ name: 'root_activity', description: 'Root user activity detected.', risk_score: 74, entity: 'user_name' }) }),
      },
    ] }),
    // Complex conditions are made possible by using the meta_condition inspector.
    sub.tf.meta.switch({ cases: [
      {
        // This condition requires both of these statements to be true:
        //
        // - The `source_ip` field is a public IP address.
        // - The `user_name` field contains either `root` or `admin`.
        condition: sub.cnd.all([
          sub.cnd.meta.condition({ condition: sub.cnd.none(
            sub.pattern.cnd.net.ip.internal(key='source_ip')
          ) }),
          sub.cnd.meta.condition({ condition: sub.cnd.any([
            sub.cnd.str.has({ obj: { src: 'user_name' }, value: 'root' }),
            sub.cnd.str.has({ obj: { src: 'user_name' }, value: 'admin' }),
          ]) }),
        ]),
        transform: sub.tf.meta.pipe({
          transforms:
            const.threat_signal({ name: 'public_ip_root_admin_activity', description: 'Public IP root or admin user activity detected.', risk_score: 99, entity: 'source_ip'
                                                                                                                                                                 + const.threat_signal({ name: 'public_ip_root_admin_activity', description: 'Public IP root or admin user activity detected.', risk_score: 99, entity: 'user_name' }) }),
        }),
      },
    ] }),
    // If the threat signal key contains matches, then those are written to the S3
    // bucket as individual events. The `auxiliary_transforms` field is used to format
    // the data as a JSON Lines file.
    //
    // If the threat signal key is empty, then no data is written to the S3 bucket.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.any([
          sub.cnd.num.len.gt({ obj: { src: const.threat_signal_key }, value: 0 }),
        ]),
        transform: sub.tf.meta.pipe({ transforms: [
          sub.tf.agg.from.array({ obj: { src: const.threat_signal_key } }),
          sub.tf.send.aws.s3(
            settings={
              bucket_name: bucket,
              file_path: { prefix: 'xdr', time_format: '2006/01/02', uuid: true },
              auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
            }
          ),
        ] }),
      },
    ] }),
  ],
}
