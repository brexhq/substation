#!/bin/sh
files=$(find . \( -name *.jsonnet -o -name *.libsonnet \))

for file in $files
do
  jsonnetfmt -i $file
done
