// This example shows how to split an OTLP/HTTP protobuf logs export into one
// message per log record and send the records to stdout. Each record keeps its
// resource and scope. Use signal 'metrics' for a /v1/metrics payload.
local sub = std.extVar('sub');

{
  tests: [
    {
      name: 'otlp',
      transforms: [
        // This is an OTLP logs export with two resources and three log records.
        sub.tf.test.message({ value: 'CpUCClMKGQoMc2VydmljZS5uYW1lEgkKB2V4YW1wbGUKIAoWZGVwbG95bWVudC5lbnZpcm9ubWVudBIGCgRwcm9kChQKCXVzZXIubmFtZRIHCgVhbGljZRK9AQoWCg1leGFtcGxlLnNjb3BlEgUxLjAuMBJoCSiwc+MCwt4YEAkqDgoMcHJvY2Vzc19leGVjMhMKCGV2ZW50LmlkEgcKBWV2dC0xMiAKFHByb2Nlc3MuY29tbWFuZF9saW5lEggKBmxzIC1sYWIUZXhhbXBsZS5wcm9jZXNzLmV4ZWMSOQkQtHPjAsLeGBANMhMKCGV2ZW50LmlkEgcKBWV2dC0yYhdleGFtcGxlLnBvbGljeS5kZWNpc2lvbgp3Ci8KGQoMc2VydmljZS5uYW1lEgkKB2V4YW1wbGUKEgoJdXNlci5uYW1lEgUKA2JvYhJEChYKDWV4YW1wbGUuc2NvcGUSBTEuMC4wEioyEwoIZXZlbnQuaWQSBwoFZXZ0LTNiE2V4YW1wbGUuYXBpLnJlcXVlc3Q=' }),
        sub.tf.fmt.from.b64(),
      ],
      // Asserts that each output message carries exactly one named log record.
      condition: sub.cnd.all([
        sub.cnd.num.len.eq({ obj: { src: 'scope_logs.0.log_records' }, value: 1 }),
        sub.cnd.num.len.gt({ obj: { src: 'scope_logs.0.log_records.0.event_name' }, value: 0 }),
      ]),
    },
  ],
  transforms: [
    // Split the export into one message per log record.
    sub.tf.format.from.otlp({ signal: 'logs' }),
    // Send the messages to stdout.
    sub.tf.send.stdout(),
  ],
}
