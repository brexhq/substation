{
  // writes objects to this S3 path: substation-example-worm/example/2022/01/01/*
  sink: {
    type: 's3',
    settings: {
      bucket: 'substation-example-worm',
      prefix: 'example',
    },
  },
  // use the transfer transform so we don't modify data in transit
  transform: {
    type: 'transfer',
  },
}
