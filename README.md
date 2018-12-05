# Gomeet

[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

The main gomeet's tools (generator, protoc plugin) and the main gomeet's library.

__WARNING__: dev in progress

## Installing

To install simply use `go get`.

```shell
$ go get -u github.com/gomeet/gomeet/gomeet
```

Or use git

```shell
git clone https://github.com/gomeet/gomeet.git $GOPATH/src/github.com/gomeet/gomeet
cd $GOPATH/src/github.com/gomeet/gomeet
make install
```

## Usage

```shell
$ gomeet help
Usage:
  gomeet [command]

Available Commands:
  help        Help about any command
  new         Create a new microservice
  version     Return version

Flags:
  -h, --help   help for gomeet

Use "gomeet [command] --help" for more information about a command.
```

## The `new` command usage

```shell
$ gomeet help new
Create a new microservice

Usage:
  gomeet new [name] [flags]

Flags:
      --cron-tasks string          Cron tasks (comma separated)
      --db-types string            DB types [mysql,postgres,postgis,sqlite,mssql] (comma separated)
      --default-port string        Default port (default "50051")
      --default-prefixes string    List of prefixes (comma separated) (default "svc-,gomeet-svc-")
      --extra-serve-flags string   extra serve flags passed to gRPC server format [<name-of-flag>@<type-of-flag[string|int]>|<flag description (no comma, no semicolon, no colon)>|<default value>] (comma separated)
      --force                      Replace files if exists
  -h, --help                       help for new
      --no-gogo                    if is true the protoc plugin is protoc-gen-go else it's protoc-gen-gogo in the Makefile file
      --proto-alias string         Protobuf pakage alias (default "pb")
      --queue-types string         Queue types [memory,rabbitmq,zeromq,sqs] (comma separated)
      --sub-services string        Sub services dependencies (comma separated)
      --ui-type string             UI type [none|simple|elm|elm-bulma|elm-milligram|elm-minimal|elm-minimal-http] (default "none")
```

## Use case

Generation of `github.com/gomeet-examples/svc-echo` service with the `gomeet new` command.

```shell
$ gomeet new github.com/gomeet-examples/svc-echo
Creating project in <YOUR_GOPATH>/github.com/gomeet-examples/svc-echo
Is this OK? [y]es/[N]o
y
2018/02/14 18:25:58 [Creating]  - <YOUR_GOPATH>/github.com/gomeet-examples/svc-echo/Gopkg.toml
...SNIP...
2018/02/14 18:25:59 [Creating]  - <YOUR_GOPATH>/github.com/gomeet-examples/svc-echo/docs/devel/add_sub_service/README.md

Print tree? [y]es/[N]o
y
.
├── third_party
│   ├── github.com
│   │   ├── gogo
│   │   │   └── protobuf
│   │   │       └── gogoproto
│   │   │           └── gogo.proto
...SNIP...

To finish project initialization do :
  $ cd <YOUR_GOPATH>/github.com/gomeet-examples/svc-echo
  $ git init
  $ git add .
  $ git commit -m 'First commit (gomeet new <YOUR_GOPATH>/github.com/gomeet-examples/svc-echo)'
  $ make tools-sync proto dep test
  $ git add .
  $ git commit -m 'Added tools and dependencies'

Do it? [y]es/[N]o
y
<YOUR_GOPATH>/github.com/gomeet-examples/svc-echo $ git init
...SNIP...

To git flow initialization do :
  $ cd <YOUR_GOPATH>/github.com/gomeet-examples/svc-echo
  $ git flow init -d

Do it? [y]es/[N]o
y
<YOUR_GOPATH>/github.com/gomeet-examples/svc-echo $ git flow init -d

Which branch should be used for bringing forth production releases?
...SNIP...

```

## TODO

- [ ] Units tests
- [ ] Add an use case see [gomeet-examples](https://github.com/gomeet-examples/)
- [x] Add ui generator
- [ ] Use bazel for build?
- [ ] Improvements
- [ ] Add `make package-<OS>-<ARCH>` directives
- [ ] Bug fix colision name for sub services definitions
- [ ] Support for flags or ENV VAR in service/service_test.go
- [ ] Make releases on github or gogs
- [ ] Add hack/run.sh on project generator

## Similar projects

- [lile](https://github.com/lileio/lile) - certainly better : less boillerplate
- [protoc-gen-gotemplate](https://github.com/moul/protoc-gen-gotemplate) - certainly better : more generic

## License

`gomeet` and `protoc-gen-gomeet-service` are released under the Apache 2.0 license. See the [LICENSE](LICENSE.txt) file for details.

