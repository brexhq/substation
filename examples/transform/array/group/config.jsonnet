// This example groups an array of arrays into an array of objects
// based on index and configured keys.
local sub = std.extVar('sub');

local files_key = 'meta files';

{
  tests: [
    {
      name: 'group',
      transforms: [
        sub.tf.test.message({ value: { file_name: ['foo.txt', 'bar.html'], file_type: ['text/plain', 'text/html'], file_size: [100, 500] } }),
        sub.tf.send.stdout(),
      ],
      // Asserts that each element in the array contains these keys:
      //  - file_name
      //  - file_type
      //  - file_size
      //  - file_extension
      condition: sub.cnd.meta.all({
        object: { source_key: '@this' },
        conditions: [
          sub.cnd.num.len.greater_than({ obj: { src: 'file_type' }, value: 0 }),
          sub.cnd.num.len.greater_than({ obj: { src: 'file_size' }, value: 0 }),
          sub.cnd.num.len.greater_than({ obj: { src: 'file_name' }, value: 0 }),
          sub.cnd.num.len.greater_than({ obj: { src: 'file_extension' }, value: 0 }),
        ],
      }),
    },
  ],
  transforms: [
    // Copy the object to metadata, where it is grouped.
    sub.tf.obj.cp({ object: { target_key: files_key } }),
    // Elements from the file_name array are transformed and derived file extensions
    // are added to a new array.
    sub.tf.meta.for_each({
      object: { source_key: files_key + '.file_name', target_key: files_key + '.file_extension' },
      transforms: [
        sub.tf.string.capture(settings={ pattern: '\\.([^\\.]+)$' }),
      ],
    }),
    // The arrays grouped into an array of objects, then copied to the message's data field.
    // For example:
    //
    // [{name: name1, type: type1, size: size1, extension: extension1}, {name: name2, type: type2, size: size2, extension: extension2}]
    sub.tf.object.cp({ object: { source_key: files_key + '|@group' } }),
    sub.tf.obj.cp({ object: { source_key: '@pretty' } }),
    sub.tf.send.stdout(),
  ],
}
