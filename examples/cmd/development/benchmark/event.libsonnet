local sub = import '../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.transform.time.now({ object: { target_key: 'now' } }),
  ],
}
