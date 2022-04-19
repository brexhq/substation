#!/bin/sh
# recursively searches for Substation config Jsonnet files and compiles them into the files' local directory
files=$(find . -name config.jsonnet)

for file in $files
do
  directory=$(echo $file | sed 's|/[^/]*$||')
  jsonnet $file > $directory/config.json
done
