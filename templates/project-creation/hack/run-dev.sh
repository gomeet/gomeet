#!/bin/sh

########
# init #
########
trap killgroup 2

SCRIPT=$(readlink -f "$0")
SCRIPTPATH=$(dirname "$SCRIPT")

GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)

killgroup(){
	echo killing...
	kill 0
}

cd $SCRIPTPATH/..
$SCRIPTPATH/../_tools/bin/fswatch --config $SCRIPTPATH/../.fsw.yml &
{{ if .HasUi }}{{ if .HasUiElm }}cd $SCRIPTPATH/../ui
$SCRIPTPATH/../_tools/bin/fswatch &{{ end }}
{{ end }}

wait
