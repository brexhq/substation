local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  concurrency: 1,
  transforms: [
    sub.tf.send.stdout(),
  ],
}
