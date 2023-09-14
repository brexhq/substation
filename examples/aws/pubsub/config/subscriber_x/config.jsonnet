local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.transform.send.stdout(),
  ]
}
