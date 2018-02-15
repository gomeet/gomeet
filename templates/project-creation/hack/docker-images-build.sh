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

if [ ! -z $(docker images --filter=reference=$DOCKER_IMAGE -q) ]; then
  echo "$DOCKER_IMAGE exist $DOCKER_IMAGE will be removed"
  docker rmi -f $DOCKER_IMAGE
  echo " -> is removed"
fi;


echo "$DOCKER_IMAGE will be created"
docker build -f Dockerfile -t $DOCKER_IMAGE .
