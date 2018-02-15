#!/bin/sh

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASE_DIR=$SCRIPTPATH/..

# METHODS=$(grep -oP "rpc \K(.*)\(.*\) returns" $BASE_DIR/{{ .GoProtoPkgAlias }}/{{ .ShortName }}.proto | sed -n -e 's/^\([[:alnum:]]\+\).*/\1/p')
METHODS=$($SCRIPTPATH/grpc-list-method.sh)

for fn in $METHODS
do
  $SCRIPTPATH/grpc-edit-method.sh $fn
done
