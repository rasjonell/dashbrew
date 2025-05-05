#!/bin/bash

count=${1:-15}

for (( i=0; i<count; i++ )); do
  echo $(( RANDOM % 101 ))
done
