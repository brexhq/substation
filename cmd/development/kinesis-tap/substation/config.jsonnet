local sub = import '../../../../substation.libsonnet';

{
  transforms: [
    sub.tf.send.stdout(),
  ],
}
