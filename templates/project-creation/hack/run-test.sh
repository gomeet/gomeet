#!/bin/sh

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")
cd $SCRIPTPATH/..
make test -s 2>/dev/null | grep -v "{{ lowerSnakeCase .Name }}:" | grep -v -e '^$'

