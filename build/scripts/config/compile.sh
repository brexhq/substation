#!/bin/sh
files=$(find . -name config.jsonnet)

for file in $files
do
  directory=$(echo $file | sed 's|/[^/]*$||')
  jsonnet $file > $directory/config.json
done
