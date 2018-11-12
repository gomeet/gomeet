# {{ .Name }} docker-compose usage

## Build docker image

```bash
make docker
--or--
docker build -t {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __) .
```

## Launch containers

```bash
make start
--or--
docker-compose up -d
```

## Stop containers

```bash
make stop
--or--
docker-compose down -d
```

## Run clients with docker

### Console

```bash
docker-compose run console-{{ .ShortName }}
```

Detach console with `Ctrl + p Ctrl + q` and attach with :

```bash
docker attach {{ lowerNospaceCase .Name }}_console-{{ .ShortName }}_X
```

### Client with docker

```bash
docker run \
    --net={{ lowerNospaceCase .Name }}_grpc \
    -it {{ .ProjectGroupName }}/{{ .Name }}:$(cat VERSION | tr +- __) cli echo 42 --address=svc:{{ .DefaultPort }}
```

### Curl with docker use gomeet/gomeet-curl

[Docker Hub](https://hub.docker.com/r/gomeet/gomeet-curl/) - [Source](https://github.com/gomeet/gomeet-curl)

```bash
# use HTTP/1.1 api
docker run \
    --net={{ lowerNospaceCase .Name }}_http \
    -it gomeet/gomeet-curl -X POST http://svc:{{ .DefaultPort }}/api/v1/echo -d '{"id": "{id}"}'

# status and metrics
docker run \
    --net={{ lowerNospaceCase .Name }}_http \
    -it gomeet/gomeet-curl http://svc-{{ .ShortName }}:{{ .DefaultPort }}/status

docker run \
    --net={{ lowerNospaceCase .Name }}_http \
    -it gomeet/gomeet-curl http://svc-{{ .ShortName }}:{{ .DefaultPort }}/metrics
```

## Grafana configuration

- [See grafana documentation](../grafana/README.md)

