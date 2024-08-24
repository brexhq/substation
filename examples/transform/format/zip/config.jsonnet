// This example shows how to unzip a file and send the contents to stdout.
// Add the two data files in this directory to a Zip file and send it to
// Substation. You can use this command to create the Zip file:
//  zip data.zip data.jsonl data.csv
local sub = import '../../../../substation.libsonnet';

{
  transforms: [
    // Unzip the file. The contents of each file in the Zip file are
    // now messages in the pipeline (including EOL characters, if any).
    sub.tf.format.from.zip(),
    // Create individual messages from the contents of each file.
    sub.tf.agg.from.string({ separator: '\n' }),
    // Send the messages to stdout.
    sub.tf.send.stdout(),
  ],
}
