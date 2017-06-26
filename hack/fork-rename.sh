#!/bin/sh -e

if [[ -z $1 || -z $2 ]]; then
  echo "Usage:  fork-rename.sh <github org> <repo name>"
  exit 1
fi

OS_TARGET=`uname -s`  
REPLACE="s#github.com/openshift/open-service-broker-sdk#github.com/$1/$2#g"

if  [ "$OS_TARGET" == "Darwin" ]; then
  find . -type f -name *.go -exec sed -i '' "${REPLACE}" {} +
  sed -i ''  "${REPLACE}" Makefile
  exit
fi
  
find . -type f -name *.go -exec sed -i "${REPLACE}" {} +
sed -i  "${REPLACE}" Makefile