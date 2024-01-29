local sub = import '../../../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.tf.send.stdout(),
  ],
}