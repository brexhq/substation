local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.transform.send.aws.kinesis_data_stream(
      settings={stream:'substation_raw'},
    )
  ]
}
