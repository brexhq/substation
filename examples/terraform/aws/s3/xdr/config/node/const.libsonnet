local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  threat_signals_key: 'threat.signals',
  // threat_signal is a custom function that appends a threat signal to an
  // event as enrichment metadata.
  //
  // An alternate approach is to compose a new threat signal event within
  // the message metadata and send it as a separate event. This results in
  // smaller events with less context and requires a correlation value
  // (e.g., hash, ID) to link the threat signal to the original event.
  threat_signal(settings): sub.tf.obj.insert({
    obj: { trg: sub.helpers.obj.append_array($.threat_signals_key) },
    value: { name: settings.name, description: settings.description, risk_score: settings.risk_score },
  }),
}
