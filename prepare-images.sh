#!/bin/bash
set -x #echo on

Images=( 'c' 'cpp' 'csharp' 'go' 'java' 'kotlin' 'python')

for i in "${Images[@]}" ; do
  echo "Started building image: ${i}"
  docker build -t "autochecker-${i}" -f "./dockerfiles/${i}/Dockerfile" .
  echo "OK"
done