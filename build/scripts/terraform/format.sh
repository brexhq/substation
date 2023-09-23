#!/bin/sh
files=$(find . -name *.tf)

for file in $files
do
  terraform fmt $file
done
