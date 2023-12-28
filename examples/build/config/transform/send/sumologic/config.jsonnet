// This example creates a newline delimited JSON (ndjson) document that can be
// sent to a Sumo Logic HTTPS endpoint.
local sub = import '../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    // Sumo Logic has a strict limit of 1MB per request.
    sub.tf.aggregate.to.array({ buffer: { size: 1024*1024 } }),
    sub.tf.array.join({ separator: '\n' }),
    sub.tf.send.http.post({
      url: 'https://endpoint6.collection.us2.sumologic.com/receiver/v1/http/xxxxxxxxxx',
      // You can override the default source category associated with the URL.
      // headers: [{key: 'X-Sumo-Category', value: 'testing/substation'}]
    }),
  ],
}
