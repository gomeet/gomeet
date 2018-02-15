#!/bin/sh

if [ "$1" = "" ]
then
  echo "usage: $0 <GrpcServiceName (in KebabCase)>"
  exit 1
fi

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
BASE_DIR=$SCRIPTPATH/..

fn=$(echo "$1" | tr '-' '_' | sed 's/./\U&/')
fn_underscore=$(echo $fn | sed 's/\([a-z0-9]\)\([A-Z]\)/\1_\L\2/g' | tr '[:upper:]' '[:lower:]')
fn=$(echo $fn_underscore | sed -r 's/(^|_)([a-z])/\U\2/g')
msg=$fn"Request"
resp=$fn"Response"

if [ "$GOMEET_EDITOR" = "" ]
then
	if [ "$EDITOR" = "" ]
	then
		EDITOR="vim"
	fi
	GOMEET_EDITOR=$EDITOR' "-c tabdo /'$fn'\|'$fn_underscore'" -p'
fi

METHOD_FILES="$BASE_DIR/{{ .GoProtoPkgAlias }}/{{ .ShortName }}.proto
$BASE_DIR/service/service.go
$BASE_DIR/cmd/root.go
$BASE_DIR/cmd/cli.go
$BASE_DIR/cmd/remotecli/cmd_help.go
$($SCRIPTPATH/grpc-method-files-list.sh $1)
$BASE_DIR/docs/grpc-services/README.md
"

eval $GOMEET_EDITOR \
	$METHOD_FILES
