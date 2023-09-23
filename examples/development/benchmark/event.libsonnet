local sub = import '../../../build/config/substation.libsonnet';

{
  transforms: [
    sub.transform.time.now(
      settings={ object: { set_key: 'now' } }
    ),
  ],
}
