#!/bin/sh

if [ "$1" = "" ]
then
  echo "usage: $0 <GrpcServiceName (in KebabCase)>"
  exit 1
fi

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASE_DIR=$SCRIPTPATH/..

cd $BASE_DIR

fn=$(echo "$1" | tr '-' '_' | sed 's/./\U&/')
fn_underscore=$(echo $fn | sed 's/\([a-z0-9]\)\([A-Z]\)/\1_\L\2/g' | tr '[:upper:]' '[:lower:]')
fn=$(echo $fn_underscore | sed -r 's/(^|_)([a-z])/\U\2/g')
msg=$fn"Request"
resp=$fn"Response"

EDITOR='vim "-c tabdo /'$fn'\|'$fn_underscore'" -p'

echo -n "Are you sure? (y/N) "
read confirm
case "$confirm" in
y|Y)
  echo "______________"
  echo "REMOVING FILES"
  echo ""
  FILES="$BASE_DIR/service/grpc_$fn_underscore.go \
  $BASE_DIR/service/grpc_"$fn_underscore"_test.go \
  $BASE_DIR/cmd/remotecli/cmd_$fn_underscore.go \
  $BASE_DIR/cmd/functest/helpers_$fn_underscore.go \
  $BASE_DIR/cmd/functest/grpc_$fn_underscore.go \
  $BASE_DIR/cmd/functest/http_$fn_underscore.go"

  for f in $FILES
  do
    f=$(readlink -f $f)
    echo "rm -rf $f"
    rm -rf $f
  done

  echo ""
  ;;
n|N|*)
  echo "Cancel..."
  exit 1
  ;;
esac
