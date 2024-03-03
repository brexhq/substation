local sub = import '../../../../../../../build/config/substation.libsonnet';

// This is a placeholder that must be replaced with the bucket produced by Terraform.
local bucket = 'substation-3e820117-61f0-2fbb-05c4-1fba0db9d82c';
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
        transform: const.threat_signal({ name: 'root_activity', description: 'Root user activity detected.', risk_score: 74 }),
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
        transform: const.threat_signal({ name: 'public_ip_root_admin_activity', description: 'Public IP root or admin user activity detected.', risk_score: 99 }),
      },
    ] }),
    // If the event contains a threat signal, then it's written to the XDR path
    // in the S3 bucket; otherwise the event is discarded. The `auxiliary_transforms`
    // field is used to format the data as a JSON Lines file.
    //
    // If there are no threat signals, then the event is discarded.
    sub.tf.meta.switch({ cases: [
      {
        condition: sub.cnd.any([
          sub.cnd.num.len.gt({ obj: { src: const.threat_signals_key }, value: 0 }),
        ]),
        transform: sub.tf.send.aws.s3(
          settings={
            bucket_name: bucket,
            file_path: { prefix: 'xdr', time_format: '2006/01/02', uuid: true, suffix: '.jsonl' },
            auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
          }
        ),
      },
    ] }),
  ],
}
