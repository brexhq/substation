// This example shows how to use the `utility_secret` transform to
// retrieve a secret and reference it in a subsequent transform.
//
// Test this example using the substation CLI:
//  SUBSTATION_EXAMPLE_URL=https://www.gutenberg.org/files/2701/old/moby10b.txt substation test config.jsonnet
local sub = std.extVar('sub');

// The secret is retrieved from the environment variable named
// `SUBSTATION_EXAMPLE_URL` and referenced in subsequent transforms using
// the ID value `ENV_VAR`.
local secret = sub.secrets.environment_variable({ id: 'ENV_VAR', name: 'SUBSTATION_EXAMPLE_URL' });

{
  tests: [
    {
      name: 'http_secret',
      transforms: [
        // An empty message is sufficient for this test.
        sub.tf.test.message(),
      ],
      condition: sub.cnd.num.len.gt({ value: 0 }),
    },
  ],
  // The `utility_secret` transform retrieves the secret from the environment
  // variable and keeps it in memory. The `enrich_http_get` transform references
  // the secret using the ID value `ENV_VAR`. In this example, the secret is the
  // URL of a web page that is retrieved by the `enrich_http_get` transform and
  // sent to stdout by the `send_stdout` transform.
  transforms: [
    sub.transform.utility.secret({ secret: secret }),
    sub.transform.enrich.http.get({ url: '${SECRET:ENV_VAR}' }),
    // Moby Dick is a large text, so the max size of the batch
    // has to be increased, otherwise the data won't fit.
    sub.tf.send.stdout({ batch: { size: 10000000 } }),
  ],
}
