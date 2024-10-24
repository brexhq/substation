// This example configures send transforms with additional transforms that
// are executed after the data is buffered and before it is sent. The
// transforms applied inside of the send transform do not affect the data
// sent through the main pipeline. All send transforms use this behavior.
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'aux_transforms',
      transforms: [
        sub.tf.test.message({ value: { a: 'b' } }),
        sub.tf.test.message({ value: { c: 'd' } }),
        sub.tf.test.message({ value: { e: 'f' } }),
        sub.tf.test.message({ value: { g: 'h' } }),
        sub.tf.test.message({ value: { i: 'j' } }),
        sub.tf.test.message({ value: { k: 'l' } }),
        sub.tf.test.message({ value: { m: 'n' } }),
        sub.tf.test.message({ value: { o: 'p' } }),
        sub.tf.test.message({ value: { q: 'r' } }),
        sub.tf.test.message({ value: { s: 't' } }),
        sub.tf.test.message({ value: { u: 'v' } }),
        sub.tf.test.message({ value: { w: 'x' } }),
        sub.tf.test.message({ value: { y: 'z' } }),
      ],
      condition: sub.cnd.num.len.gt({ value: 0 }),
    },
  ],
  transforms: [
    // By default all data is buffered before it is sent.
    sub.tf.send.stdout({
      // auxiliary_transforms is a sub-pipeline executed after the data is
      // batched and before it is sent. The data is scoped to the send transform
      // and results are not forwarded to the next transform in the pipeline.
      // Any transform can be used here, including additional send transforms.
      //
      // If auxiliary_transforms is not used, then the batched data is sent individually
      // without modification.
      auxiliary_transforms: [
        sub.tf.object.insert({ object: { target_key: 'transformed_by' }, value: 'send_stdout' }),
      ],
    }),
    // By default, send.file writes data to `$(pwd)/[year]/[month]/[day]/[uuid]`.
    sub.tf.send.file({
      // This sub-pipeline creates a newline delimited JSON (NDJSON) file. Uncomment
      // the additional transforms to compress and encode the file.
      aux_tforms: [
        sub.tf.object.insert({ object: { target_key: 'transformed_by' }, value: 'send_file' }),
        sub.tf.agg.to.string({ separator: '\n' }),
        sub.tf.str.append({ suffix: '\n' }),
      ],
    }),
    // This transform is included to show that the data is not modified outside of
    // any individual transform's scope. Since this transform has a low buffer count,
    // most data is sent to stdout before the data from any previous transform is.
    sub.tf.send.stdout({ batch: { count: 1 } }),
  ],
}
