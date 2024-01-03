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
    sub.tf.obj.cp(settings={ object: { source_key: '[file_names,file_types,file_sizes]', target_key: files_key } }),
    sub.tf.send.stdout(),
    // Elements of the file_names array are transformed, the file extension
    // results are appended to the files array, and the arrays are zipped together.
    // For example:
    //
    // [[name1, type1, size1, extension1], [name2, type2, size2, extension2]]
    sub.tf.meta.for_each({
      object: { source_key: 'file_names', target_key: sub.helpers.object.append_array(files_key) },
      transform: sub.tf.string.capture(settings={ pattern: '\\.([^\\.]+)$' }),
    }),
    sub.tf.array.zip({ object: { source_key: files_key, target_key: files_key } }),
    sub.tf.send.stdout(),
    // The array of arrays is transformed into an array of objects based on
    // index and configured keys. For example:
    //
    // [{name: name1, type: type1, size: size1, extension: extension1}, {name: name2, type: type2, size: size2, extension: extension2}]
    sub.tf.array.to.object({ object: { source_key: files_key, target_key: files_key }, object_keys: ['name', 'type', 'size', 'extension'] }),
    sub.tf.send.stdout(),
    // The array of objects are transformed into new events.
    sub.tf.obj.cp({ object: { source_key: files_key } }),
    sub.tf.agg.from.array(),
    sub.tf.send.stdout(),
  ],
}
