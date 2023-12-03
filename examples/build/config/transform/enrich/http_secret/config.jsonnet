// This example shows how to use the `utility_secret` transform to
// retrieve a secret and reference it in a subsequent transform.
local sub = import '../../../../../../build/config/substation.libsonnet';

// The secret is retrieved from the environment variable named
// `SUBSTATION_EXAMPLE_URL` and referenced in subsequent transforms using
// the ID value `ENV_VAR`.
//
// Run this on the local system as an example:
//  export SUBSTATION_EXAMPLE_URL=https://www.gutenberg.org/files/2701/old/moby10b.txt
local secret = sub.secrets.environment_variable(
  settings={ id: 'ENV_VAR', name: 'SUBSTATION_EXAMPLE_URL' }
);

{
  // The `utility_secret` transform retrieves the secret from the environment
  // variable and keeps it in memory. The `enrich_http_get` transform references
  // the secret using the ID value `ENV_VAR`. In this example, the secret is the
  // URL of a web page that is retrieved by the `enrich_http_get` transform and
  // sent to stdout by the `send_stdout` transform.
  transforms: [
    sub.transform.utility.secret(
      settings={ secret: secret }
    ),
    sub.transform.enrich.http.get(
      settings={ url: '${SECRET:ENV_VAR}' }
    ),
    sub.tf.send.stdout(),
  ],
}
