local sub = import '../../build/config/substation.libsonnet';

{
  transforms: [
    sub.interfaces.transform.send.stdout,
  ]
}
