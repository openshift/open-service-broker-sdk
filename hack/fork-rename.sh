#!/bin/sh -e

if [[ -z $1 || -z $2 ]]; then
  echo "Usage:  fork-rename.sh <github org> <repo name>"
  exit 1
fi

OS_TARGET=`uname -s`  
BACKUP=""
if  [ "$OS_TARGET" == "Darwin" ]; then
  BACKUP="''"
fi
  
find . -type f -name *.go -exec sed -i $BACKUP  "s#github.com/openshift/open-service-broker-sdk#github.com/$1/$2#g" {} +
  sed -i $BACKUP  "s#github.com/openshift/open-service-broker-sdk#github.com/$1/$2#g" Makefile