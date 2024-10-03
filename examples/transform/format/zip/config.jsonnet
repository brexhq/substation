// This example shows how to unzip a file and send the contents to stdout.
// Add the two data files in this directory to a Zip file and send it to
// Substation. You can use this command to create the Zip file:
//  zip data.zip data.jsonl data.csv
local sub = import '../../../../substation.libsonnet';

{
  tests: [
    {
      name: 'zip',
      transforms: [
        // This is a Zip file containing a CSV file and a JSONL file.
        sub.tf.test.message({ value: 'UEsDBAoAAAAAAMeuFFmAS9oWGgAAABoAAAAIABwAZGF0YS5jc3ZVVAkAA/VyxWb1csVmdXgLAAEE9gEAAAQUAAAAZm9vLGJhcgpiYXoscXV4CnF1dXgsY29yZ2VQSwMEFAAAAAgAx64UWYViom8nAAAALAAAAAoAHABkYXRhLmpzb25sVVQJAAP1csVm9XLFZnV4CwABBPYBAAAEFAAAAKtWSsvPV7JSSkosUqrlqgbSVUBeYWkFmFdYCmRYKSXnF6WnKtUCAFBLAQIeAwoAAAAAAMeuFFmAS9oWGgAAABoAAAAIABgAAAAAAAEAAACkgQAAAABkYXRhLmNzdlVUBQAD9XLFZnV4CwABBPYBAAAEFAAAAFBLAQIeAxQAAAAIAMeuFFmFYqJvJwAAACwAAAAKABgAAAAAAAEAAACkgVwAAABkYXRhLmpzb25sVVQFAAP1csVmdXgLAAEE9gEAAAQUAAAAUEsFBgAAAAACAAIAngAAAMcAAAAAAA==' }),
        sub.tf.fmt.from.b64(),
      ],
      // Asserts that each message is not empty.
      condition: sub.cnd.num.len.gt({ value: 0 }),
    }
  ],
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
