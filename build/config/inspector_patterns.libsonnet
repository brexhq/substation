local inspector = import 'inspector.libsonnet';

{
  ip: {
    // checks if an IP address is private.
    //
    // use with the ANY operator to match private IP addresses.
    // use with the NONE operator to match public IP addresses.
    private(key=''): [
      inspector.inspect(inspector.ip(type='loopback'), key=key),
      inspector.inspect(inspector.ip(type='multicast'), key=key),
      inspector.inspect(inspector.ip(type='multicast_link_local'), key=key),
      inspector.inspect(inspector.ip(type='private'), key=key),
      inspector.inspect(inspector.ip(type='unicast_link_local'), key=key),
      inspector.inspect(inspector.ip(type='unspecified'), key=key),
    ],
  },
  length: {
    // checks if data is equal to zero.
    //
    // use with the ANY / ALL operator to match empty data.
    // use with the NONE operator to match non-empty data.
    eq_zero(key=''): inspector.inspect(inspector.length(type='equals', value=0), key=key),
    // checks if data is greater than zero.
    //
    // use with the ANY / ALL operator to match non-empty data.
    // use with the NONE operator to match empty data.
    gt_zero(key=''): inspector.inspect(inspector.length(type='greater_than', value=0), key=key),
  },
}
