#!/usr/bin/env bash

#  ./verify.sh target  
#  ./verify.sh .

DIR=$1
for f in $(find $DIR -name "*.lql")
do
  #echo "lql found: $f"
  lqlp $f
done