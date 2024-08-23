// This example shows how to make scan requests and retrieve
// results using the urlscan API (https://urlscan.io/docs/api/).
local sub = import '../../../../substation.libsonnet';

local headers = { 'API-Key': '${SECRET:URLSCAN}', 'Content-Type': 'application/json' };

{
  transforms: [
    // Retrieve the urlscan API key from the secrets store.
    // (Never put a secret directly into a configuration.)
    sub.transform.utility.secret({
      // The API key is stored in an environment variable named
      // `URLSCAN_API_KEY`.
      secret: sub.secrets.environment_variable({ id: 'URLSCAN', name: 'URLSCAN_API_KEY' }),
    }),
    // Sends a scan request and waits for the result. This
    // follows recommended practices from the urlscan API docs,
    // and will try to fetch the result up to 3 times over 15s.
    // If there are no results after retrying, then the unmodified
    // message is sent to stdout.
    sub.tf.enrich.http.post({
      object: { body_key: '@this', target_key: 'meta response' },
      url: 'https://urlscan.io/api/v1/scan/',
      headers: headers,
    }),
    sub.tf.util.delay({ duration: '5s' }),
    sub.tf.meta.err({
      error_messages: ['retry limit reached'],  // Errors are caught in case the retry limit is reached.
      transforms: [
        sub.tf.meta.retry({
          // This condition runs on the result of the transforms. If
          // it returns false, then the transforms are retried until
          // it returns true or the retry settings are exhausted.
          condition: sub.cnd.all([
            sub.cnd.num.len.gt({ object: { source_key: 'meta result.task.time' }, value: 0 }),
          ]),
          transforms: [
            sub.tf.enrich.http.get({
              object: { source_key: 'meta response.uuid', target_key: 'meta result' },
              url: 'https://urlscan.io/api/v1/result/${DATA}',  // DATA is the value of the source_key.
              headers: headers,
            }),
          ],
          retry: { delay: '5s', count: 3 },  // Retry up to 3 times with a 5 second delay (5s, 5s, 5s).
        }),
      ],
    }),
    sub.tf.obj.cp({ object: { source_key: 'meta result' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout({ batch: { size: 1000 * 1000 * 5 } }),  // 5MB (the results can be large).
  ],
}
