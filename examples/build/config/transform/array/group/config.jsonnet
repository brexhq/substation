// This example groups an array of arrays into an array of objects
// based on index and configured keys.
local sub = import '../../../../../../build/config/substation.libsonnet';

local files_key = 'files';

{
  concurrency: 1,
  transforms: [
    // This example sends data to stdout at each step to iteratively show
    // how the data is transformed.
    sub.tf.send.stdout(),
    // Each array is appended to the files array. This creates
    // an array of arrays. For example:
    //
    // [[name1, name2], [type1, type2], [size1, size2]]
    sub.tf.obj.cp(settings={ object: { key: '[file_names,file_types,file_sizes]', set_key: files_key } }),
    sub.tf.send.stdout(),
    // Elements of the file_names array are transformed and the file extension
    // results are appended to the files array. For example:
    //
    // [[name1, name2], [type1, type2], [size1, size2], [extension1, extension2]]
    sub.tf.meta.for_each(settings={
      object: { key: 'file_names', set_key: sub.helpers.key.append_array(files_key) },
      transform: sub.tf.string.capture(settings={ pattern: '\\.([^\\.]+)$' }),
    }),
    sub.tf.send.stdout(),
    // The array of arrays is transformed into an array of objects based on
    // index and configured keys. For example:
    //
    // [{name: name1, type: type1, size: size1}, {name: name2, type: type2, size: size2}]
    sub.tf.arr.group(settings={ object: { key: files_key, set_key: files_key }, group_keys: ['name', 'type', 'size', 'extension'] }),
    sub.tf.send.stdout(),
    // The array of objects are transformed into new events.
    sub.tf.obj.cp(settings={ object: { key: files_key } }),
    sub.tf.agg.from.array(),
    sub.tf.send.stdout(),
  ],
}
