// This example creates a newline delimited JSON (ndjson) document that can be
// sent to a Sumo Logic HTTPS endpoint.
//
// More information about Sumo Logic HTTP upload can be found here:
// https://help.sumologic.com/docs/send-data/hosted-collectors/http-source/logs-metrics/upload-logs/
local sub = import '../../../../../build/config/substation.libsonnet';

// Sumo Logic has a strict limit of 1MB per request.
local max_size = 1000 * 1000;

{
  concurrency: 1,
  transforms: [
    sub.tf.send.http.post({
      batch: { size: max_size },
      auxiliary_transforms: sub.pattern.tf.fmt.jsonl,
      // There is no authentication, so the URL should be treated like a secret.
      url: 'https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/xxxxxxxxxx',
      // You can override the default source category associated with the URL.
      // headers: [{key: 'X-Sumo-Category', value: 'testing/substation'}]
    }),
  ],
}
