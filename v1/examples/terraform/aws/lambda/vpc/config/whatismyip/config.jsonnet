local sub = import '../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // Get the IP address of the service and return it in response.
    sub.tf.enrich.http.get(settings={ url: 'https://ipinfo.io/ip' }),
    sub.tf.object.copy(
      settings={ object: { target_key: 'ip' } },
    ),
  ],
}
