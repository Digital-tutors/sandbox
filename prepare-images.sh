#!/bin/bash
set -x #echo on

Images=( 'clang' 'cpp' 'csharp' 'golang' 'java' 'kotlin' 'student-python')

for i in "${Images[@]}" ; do
  echo "Started building image: ${i}"
  docker build -t "autochecker-${i}" -f "./dockerfiles/${i}/Dockerfile" .
  echo "OK"
done