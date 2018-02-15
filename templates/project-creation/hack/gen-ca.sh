#!/bin/bash

if [ "$1" = "" ]
then
    echo "usage: $0 ca_name"
    exit 1
fi

CANAME=$1

# Define base directory
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  TARGET="$(readlink "$SOURCE")"
  if [[ $TARGET == /* ]]; then
    echo "SOURCE '$SOURCE' is an absolute symlink to '$TARGET'"
    SOURCE="$TARGET"
  else
    DIR="$( dirname "$SOURCE" )"
    echo "SOURCE '$SOURCE' is a relative symlink to '$TARGET' (relative to '$DIR')"
    SOURCE="$DIR/$TARGET" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
  fi
done
RDIR="$( dirname "$SOURCE" )"
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

cd $DIR/../

if ! [ -x "$(command -v _tools/bin/retool)" ] || ! [ -x "$(command -v _tools/bin/certstrap)" ]; then
    echo "The retool command is unavailable."
    make tools
    if ! [ -x "$(command -v _tools/bin/retool)" ] || ! [ -x "$(command -v _tools/bin/certstrap)" ]; then
        echo "The retool or certstrap command is unavailable and make tools doesn't install it"
        echo "Check Makefile"
        exit 1
    fi
fi

mkdir -p data/certs

_tools/bin/retool do certstrap --depot-path "data/certs" init --common-name $CANAME
exit 0
