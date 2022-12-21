local condition = import 'condition.libsonnet';

{
  ip: {
    // evalutes if an IP address is private.
    //
    // combine with the OR operator to match private IP addresses:
    // operatorPatterns.or(inspectorPatterns.ip.private('addr'))
    //
    // combine with the NAND operator to match public IP addresses:
    // operatorPatterns.nand(inspectorPatterns.ip.private('addr'))
    private(addr=''): [
      condition.inspector(condition.ip(type='loopback'), key=addr),
      condition.inspector(condition.ip(type='multicast'), key=addr),
      condition.inspector(condition.ip(type='multicast_link_local'), key=addr),
      condition.inspector(condition.ip(type='private'), key=addr),
      condition.inspector(condition.ip(type='unicast_link_local'), key=addr),
      condition.inspector(condition.ip(type='unspecified'), key=addr),
    ],
    public(addr=''): {

    },
  },
  length: {
    eq_zero: condition.length(type='equals', value=0),
    // evalutes if data is greater than zero.
    //
    // use with the AND operator to match non-empty data.
    gt_zero: condition.length(type='greater_than', value=0),
  },
}
