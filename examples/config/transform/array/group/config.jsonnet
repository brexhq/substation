// This example groups an array of arrays into an array of objects
// based on index and configured keys.
local sub = import '../../../../../build/config/substation.libsonnet';

local files_key = 'meta files';

{
  concurrency: 1,
  transforms: [
    // This example sends data to stdout at each step to iteratively show
    // how the data is transformed.
    sub.tf.send.stdout(),
    // Copy the object to metadata, where it is grouped.
    sub.tf.obj.cp({ object: { target_key: files_key } }),
    // Elements from the file_name array are transformed and derived file extensions
    // are added to a new array.
    sub.tf.meta.for_each({
      object: { source_key: sub.helpers.object.get_element(files_key, 'file_name'), target_key: sub.helpers.object.append(files_key, 'file_extension') },
      transform: sub.tf.string.capture(settings={ pattern: '\\.([^\\.]+)$' }),
    }),
    // The arrays grouped into an array of objects, then copied to the message's data field.
    // For example:
    //
    // [{name: name1, type: type1, size: size1, extension: extension1}, {name: name2, type: type2, size: size2, extension: extension2}]
    sub.tf.object.cp({ object: { source_key: files_key + '|@group' } }),
    sub.tf.send.stdout(),
  ],
}
