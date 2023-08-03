local sub = import '../../build/config/substation.libsonnet';

{
  transforms: [
    sub.interfaces.transform.proc_copy(
      settings={ key: 'foo', set_key: 'bar' },
    ),
    sub.interfaces.transform.proc_insert(
      settings={ set_key: 'qux', value: 'quux' },
    ),
    sub.interfaces.transform.send_stdout,
  ]
}
