#!/bin/sh
files=$(find . -name "*.jsonnet")

for file in $files
do
  # 'rev | cut | rev' converts "path/to/file.jsonnet" to "path/to/file.json"
  f=$(echo $file | rev | cut -c 4- | rev)
  # This is run from the root of the repo.
  jsonnet --ext-code-file sub="./substation.libsonnet" $file > $f
done
