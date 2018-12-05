#!/bin/sh

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASE_DIR=$SCRIPTPATH/..

METHODS=$($SCRIPTPATH/grpc-list-method.sh)

for fn in $METHODS
do
  $SCRIPTPATH/grpc-edit-method.sh $fn
done
