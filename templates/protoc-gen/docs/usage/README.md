# {{ .Name }} usage

## Basic usage

- Run server

```shell
{{ .Name }} serve --address <server-address>

# serve gRPC and HTTP multiplexed on localhost:3000
{{ .Name }} serve --address localhost:3000

# serve gRPC on localhost:3000 and HTTP on localhost:3001
{{ .Name }} serve --grpc-address localhost:3000 --http-address localhost:3001

# more info
{{ .Name }} help serve
```

- Run cli client

  ```shell
{{ cliCmdHelpString .Name .ProtoFiles }}
  $ {{ .Name }} cli --address localhost:42000 version

  # more info
  {{ .Name }} help cli
  ```

- Run console client

```shell
$ {{ .Name }} console --address=localhost:3000
INFO[0000] {{ .Name }} console  Exit=exit HistoryFile="/tmp/{{ .Name }}-62852.tmp" Interrupt="^C"
└─┤{{ .Name }}-0.1.8+dev@localhost:{{ .DefaultPort }}├─$ help
INFO[0002] HELP :
{{ remoteCliHelp .Name .ProtoFiles }}
	┌─ service_address
	└─ return service address

	┌─ jwt [<token>]
	└─ display current jwt or save none if it's set

	┌─ console_version
	└─ return console version

	┌─ tls_config
	└─ display TLS client configuration

	┌─ help
	└─ display this help

	┌─ exit
	└─ exit the console


└─┤{{ .Name }}-0.1.8+dev@localhost:{{ .DefaultPort }}├─$ unknow
WARN[0003] Bad arguments : "unknow" unknow
└─┤{{ .Name }}-0.1.8+dev@localhost:{{ .DefaultPort }}├─$ exit
```

- HTTP/1.1 usage (with curl):

  ```shell
{{ curlCmdHelpString .DefaultPort .Name .ProtoFiles }}
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/version
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/metrics
  $ curl -X GET    http://localhost:{{ .DefaultPort }}/status
  $ curl -X GET    http://localhost:42000/version
  ```

- Get help

```shell
{{ .Name }} help

# or get help directly for a command
{{ .Name }} help <command[serve|cli|console]>
```

## Tests

- Use make directive

```shell
make test
```

- Unit tests

```shell
cd service
go test
```

- Functional tests (with an embedded server)

```shell
{{ .Name }} functest -e
```

- Load tests

```shell
{{ .Name }} loadtest --address <multiplexed server address> -n <number of sessions> -s <concurrency level>
```

## TLS authentication (via a public certificat)

- Run the server behind a well configured proxy with the credentials (cf. [nginx-example.conf](../../infra/nginx/nginx-example.conf))

- Run the clients with their TLS flag:

```shell
{{ .Name }} cli <grpc_service> <params...> \
    --address localhost:3000 \
    --tls

{{ .Name }} console \
    --address localhost:3000 \
    --tls
```

## Mutual TLS authentication

- Create a Certificate Authority:

```shell
hack/gen-ca.sh {{ lowerSnakeCase .ProjectGroupName }}_ca
ls data/certs
```

- Create two key pairs with the common name "localhost":

```shell
hack/gen-cert.sh server {{ lowerSnakeCase .ProjectGroupName }}_ca
./gencert.sh client {{ lowerSnakeCase .ProjectGroupName }}_ca
ls data/certs
```

- Run the server with its TLS credentials:

```shell
{{ .Name }} serve \
    --address localhost:3000 \
    --ca data/certs/{{ lowerSnakeCase .ProjectGroupName }}_ca.crt \
    --cert data/certs/server.crt \
    --key data/certs/server.key
```

- Run the clients with their TLS credentials:

```shell
{{ .Name }} cli <grpc_service> <params...> \
    --address localhost:3000 \
    --ca data/certs/{{ lowerSnakeCase .ProjectGroupName }}_ca.crt \
    --cert data/certs/client.crt \
    --key data/certs/client.key

{{ .Name }} console \
    --address localhost:3000 \
    --ca data/certs/{{ lowerSnakeCase .ProjectGroupName }}_ca.crt \
    --cert data/certs/client.crt \
    --key data/certs/client.key
```

## JSON Web Token support

JSON Web Token validation can be enabled on the server by providing a secret key:

```shell
{{ .Name }} serve --jwt-secret foobar
```

The token subcommand is used to generate a JWT from the secret key:

```shell
{{ .Name }} token --secret-key foobar
```

Then the cli and console subcommands can use the generated token for authentication against the JWT-enabled server:

```shell
{{ .Name }} cli --jwt <generated token> <grpc_service> <params...>
{{ .Name }} console --jwt <generated token>
```

JWT validation can be tested on the HTTP/1.1 endpoints by providing the bearer token in the "Authorization" HTTP header:

```shell
TOKEN=`{{ .Name }} token --secret-key foobar`
curl -H "Authorization: Bearer $TOKEN" -X <HTTP_VERB> http://localhost:{{ .DefaultPort }}/api/v1/<grpc_service> -d '<HTTP_REQUEST_BODY json format>'
```


