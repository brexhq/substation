local sub = import '../../../../../../build/config/substation.libsonnet';

{
  threat_signals_key: 'threat.signals',
  // threat_signal is a custom function that appends threat info to an
  // event as enrichment metadata.
  //
  // If a smaller event is needed, then the enriched threat signal can
  // be emitted as a separate event. This is similar to the implementation
  // seen in the enrichment Lambda function.
  threat_signal(settings): sub.tf.obj.insert({
    obj: { trg: sub.helpers.obj.append_array($.threat_signals_key) },
    value: { name: settings.name, description: settings.description, references: settings.references, risk_score: settings.risk_score },
  }),
}
