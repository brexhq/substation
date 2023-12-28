local sub = import '../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.transform.time.now({ obj: { dst: 'now' } }),
  ],
}
