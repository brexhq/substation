local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.interfaces.transform.send.aws_kinesis(
      settings={stream:'substation_raw'},
    )
  ]
}
