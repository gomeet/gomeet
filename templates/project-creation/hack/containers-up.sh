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
DOCKER_NETWORK="network-grpc-{{ .ProjectGroupName }}"

if [ -z $(docker network ls --filter name=$DOCKER_NETWORK -q) ]; then
  echo "$DOCKER_NETWORK not found $DOCKER_NETWORK will be created"
  docker network create $DOCKER_NETWORK
  echo " -> is created"
fi;

if [ ! -z $(docker ps --filter name=$DOCKER_SVC_CONTAINER -q) ]; then
  echo "$DOCKER_SVC_CONTAINER is started $DOCKER_SVC_CONTAINER will be stopped"
  docker stop $DOCKER_SVC_CONTAINER
  echo " -> is stopped"
fi;

if [ ! -z $(docker ps --filter name=$DOCKER_SVC_CONTAINER -q -a) ]; then
  echo "$DOCKER_SVC_CONTAINER exists $DOCKER_SVC_CONTAINER will be removed"
  docker rm $DOCKER_SVC_CONTAINER
  echo " -> is removed"
fi;

docker run -d \
  --net=$DOCKER_NETWORK \
  --name=$DOCKER_SVC_CONTAINER \
  -d \
  --restart always \
  -it $DOCKER_IMAGE
echo " -> is created"


if [ ! -z $(docker ps --filter name=$DOCKER_CONSOLE_CONTAINER -q) ]; then
  echo "$DOCKER_CONSOLE_CONTAINER is started $DOCKER_CONSOLE_CONTAINER will be stopped"
  docker stop $DOCKER_CONSOLE_CONTAINER
  echo " -> is stopped"
fi;

if [ ! -z $(docker ps --filter name=$DOCKER_CONSOLE_CONTAINER -q -a) ]; then
  echo "$DOCKER_CONSOLE_CONTAINER exists $DOCKER_CONSOLE_CONTAINER will be removed"
  docker rm $DOCKER_CONSOLE_CONTAINER
  echo " -> is removed"
fi;

docker run -d \
  --net=$DOCKER_NETWORK \
  --name=$DOCKER_CONSOLE_CONTAINER \
  -d \
  --restart always \
  -it $DOCKER_IMAGE \
  console --address=$DOCKER_SVC_CONTAINER:{{ .DefaultPort }}
echo " -> is created"


echo "
you can attach an docker on console with:

    $ docker attach $DOCKER_CONSOLE_CONTAINER

Don't forget : to detach (Ctrl + p Ctrl + q).
"
