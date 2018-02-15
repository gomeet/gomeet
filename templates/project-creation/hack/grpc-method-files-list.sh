#!/bin/sh

if [ "$1" = "" ]
then
  echo "usage: $0 <GrpcServiceName (in KebabCase)>"
  exit 1
fi

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASE_DIR=$(readlink -f $SCRIPTPATH/..)

fn=$(echo "$1" | tr '-' '_' | sed 's/./\U&/')
fn_underscore=$(echo $fn | sed 's/\([a-z0-9]\)\([A-Z]\)/\1_\L\2/g' | tr '[:upper:]' '[:lower:]')
fn=$(echo $fn_underscore | sed -r 's/(^|_)([a-z])/\U\2/g')

echo $BASE_DIR/service/grpc_"$fn_underscore".go
echo $BASE_DIR/cmd/remotecli/cmd_"$fn_underscore".go
echo $BASE_DIR/service/grpc_"$fn_underscore"_test.go
echo $BASE_DIR/cmd/functest/helpers_"$fn_underscore".go
echo $BASE_DIR/cmd/functest/grpc_"$fn_underscore".go
echo $BASE_DIR/cmd/functest/http_"$fn_underscore".go

