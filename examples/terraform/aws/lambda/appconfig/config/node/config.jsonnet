local sub = import '../../../../../../../../build/config/substation.libsonnet';

{
  transforms: [
    // This will always fail validation because the settings are invalid.
    sub.tf.object.delete(
      settings={ object: { missing_key: 'abc' } }
    ),
  ],
}
