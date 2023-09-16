#!/bin/sh
files=$(find . -name *.jsonnet)

for file in $files
do
  # 'rev | cut | rev' converts "path/to/file.jsonnet" to "path/to/file.json"
  f=$(echo $file | rev | cut -c 4- | rev)
  jsonnet $file > $f
done
