#!/bin/bash
set -e

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

VERSION=$(cat $DIR/../VERSION | tr +- __)
DOCKER_IMAGE="{{ .ProjectGroupName }}/{{ .Prefix }}{{ .ShortName }}:$VERSION"

DOCKER_SVC_CONTAINER="svc-{{ .ShortName }}"
DOCKER_CONSOLE_CONTAINER="console-{{ .ShortName }}"

if [ ! -z $(docker ps --filter name=$DOCKER_SVC_CONTAINER -q) ]; then
  echo "$DOCKER_SVC_CONTAINER is started $DOCKER_SVC_CONTAINER will be stopped"
  docker stop $DOCKER_SVC_CONTAINER
  echo " -> is stopped"
else
  echo "$DOCKER_SVC_CONTAINER is not started"
fi;

if [ ! -z $(docker ps --filter name=$DOCKER_CONSOLE_CONTAINER -q) ]; then
  echo "$DOCKER_CONSOLE_CONTAINER is started $DOCKER_CONSOLE_CONTAINER will be stopped"
  docker stop $DOCKER_CONSOLE_CONTAINER
  echo " -> is stopped"
else
  echo "$DOCKER_CONSOLE_CONTAINER is not started"
fi;
