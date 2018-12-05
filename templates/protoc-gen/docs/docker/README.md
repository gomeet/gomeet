# {{ .Name }} docker usage

## Build docker image

### Regular Dockerfile

```shell
make docker
--or--
docker build -t {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __) .
```

## Use port binding on host

### 1. Launch server container

```shell
docker run -d \
    --net=network-grpc-{{ .ProjectGroupName }} \
    -p {{ .DefaultPort }}:{{ .DefaultPort }} \
    --name=svc-{{ .Name }}-1 \
    -it {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __)
```

### 2. Use client on host

- Build and use cli tool

  ```shell
  $ make
  $ cd _build
{{ cliCmdHelpString .Name .ProtoFiles }}
  $ {{ .Name }} cli --address localhost:42000 version

  # more info
  {{ .Name }} help cli
  ```

- Or use HTTP/1.1 api

  ```shell
{{ curlCmdHelpString .DefaultPort .Name .ProtoFiles }}
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/version
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/metrics
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/status
  $ curl -X GET    http://localhost:42000/version
  ```

## Do not use port binding

### 1. Create a docker's network

```shell
docker network create \
    --driver bridge network-grpc-{{ .ProjectGroupName }} &> /dev/null
```

### 2. Run server container with the previous created network

```shell
docker run -d \
    --net=network-grpc-{{ .ProjectGroupName }} \
    --name=svc-{{ .Name }} \
    -it {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __)
```

### 3. Run clients with docker

#### Console

```shell
docker run -d \
    --net=network-grpc-{{ .ProjectGroupName }} \
    --name=console-{{ .Name }} \
    -it {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __) console --address=svc-{{ .Name }}:{{ .DefaultPort }}
```

Detach console with `Ctrl + p Ctrl + q` and attach with :

```shell
docker attach console-{{ .Name }}
```

#### Client with docker

```shell
docker run \
    --net=network-grpc-{{ .ProjectGroupName }} \
    -it {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __) cli --address=svc-{{ .Name }}:{{ .DefaultPort }} <grpc_service> <params...>
```

#### Curl with docker use gomeet/gomeet-curl

[Docker Hub](https://hub.docker.com/r/gomeet/gomeet-curl/) - [Source](https://github.com/gomeet/gomeet-curl)

```shell
# use HTTP/1.1 api
docker run \
    --net=network-grpc-{{ .ProjectGroupName }} \
    -it gomeet/gomeet-curl -X POST http://svc:{{ .DefaultPort }}/api/v1/-X <HTTP_VERB> http://localhost:{{ .DefaultPort }}/api/v1/<grpc_service> -d '<HTTP_REQUEST_BODY json format>'

# status and metrics
docker run \
    --net=network-grpc-{{ .ProjectGroupName }} \
    -it gomeet/gomeet-curl http://svc-{{ .Name }}:{{ .DefaultPort }}/status

docker run \
    --net=network-grpc-{{ .ProjectGroupName }} \
    -it gomeet/gomeet-curl http://svc-{{ .Name }}:{{ .DefaultPort }}/metrics
```
