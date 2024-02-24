local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  threat_signal_key: 'meta threat.signals',
  // threat_signal is a custom function that generates a new threat signal
  // and stores it in an array. This array is later processed into new events
  // that are sent to the S3 bucket.
  //
  // An alternate approach is to add the threat signal into the original event
  // as enrichment metadata. This approach provides more context but results in
  // larger events.
  threat_signal(settings): [
    sub.tf.obj.insert({
      obj: { trg: 'meta threat' },
      value: { signal: { name: settings.name, description: settings.description, risk_score: settings.risk_score } },
    }),
    sub.tf.obj.cp({ src: settings.entity, trg: 'meta threat.signal.entity' }),
    sub.tf.obj.cp({ src: 'meta threat.signal', trg: sub.helpers.obj.append_array($.threat_signal_key) }),
  ],
}
