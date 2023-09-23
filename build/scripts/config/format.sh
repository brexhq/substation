#!/bin/sh
files=$(find . -name *.jsonnet)

for file in $files
do
  jsonnetfmt -i $file
done
