#!/bin/sh

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASE_DIR=$SCRIPTPATH/..

# METHODS=$(grep -oP "rpc \K(.*)\(.*\) returns" $BASE_DIR/pb/search-criteria.proto | sed -n -e 's/^\([[:alnum:]]\+\).*/\1/p')
METHODS=$($SCRIPTPATH/grpc-list-method.sh)

for fn in $METHODS
do
  echo "Remove $fn ?"
  $SCRIPTPATH/grpc-rm-method.sh $fn
done

