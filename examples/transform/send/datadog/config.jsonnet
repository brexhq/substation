// This example transforms data into a Datadog HTTP intake compatible format
// and sends it to Datadog using the Logs API.
//
// More information about the Datadog Logs API can be found here:
// https://docs.datadoghq.com/api/latest/logs/#send-logs
local sub = import '../../../../substation.libsonnet';

// Datadog has a strict limit of 5MB per payload. Any individual event
// larger than 1MB will be truncated on ingest.
local max_size = 1000 * 1000 * 5;

// Datadog has a strict limit of 1000 events per payload.
local max_count = 1000;

{
  transforms: [
    // Connections to the Datadog Logs API are authenticated using an API key.
    sub.transform.utility.secret({ secret: sub.secrets.environment_variable({ id: 'DD', name: 'DATADOG_API_KEY' }) }),
    sub.tf.send.http.post({
      batch: { size: max_size, count: max_count },
      auxiliary_transforms: [
        sub.tf.agg.to.array({ object: { target_key: 'message' } }),
      ],
      url: 'https://http-intake.logs.datadoghq.com/api/v2/logs',
      headers: [
        {
          key: 'DD-API-KEY',
          value: '${SECRET:DD}',
        },
        {
          key: 'ddsource',
          value: 'my-source',
        },
        {
          key: 'service',
          value: 'my-service',
        },
      ],
    }),
  ],
}
