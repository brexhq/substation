//  This example shows usage of the `network.ip` conditions.
local sub = import '../../../substation.libsonnet';

{
  tests: [
    {
      name: 'network.ip.unique_local_address',
      transforms: [
        // This is a unique local address (ULA).
        // https://en.wikipedia.org/wiki/Unique_local_address
        sub.tf.test.message({ value: "fd12:3456:789a:1::1" } ),
      ],
      // Asserts that the IP address is a unique local address.
      condition: sub.condition.network.ip.unique_local_address(),
    }
  ],
  transforms: [
    sub.tf.send.stdout(),
  ],
}
