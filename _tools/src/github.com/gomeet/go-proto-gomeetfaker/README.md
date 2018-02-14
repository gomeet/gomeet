# Golang ProtoBuf Faker Compiler

[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A `protoc` plugin that generates `New<MessageName>GomeetFaker() *<MessageName>` functions.

## Paint me a code picture

Let's take the following `proto3` snippet:

```proto
syntax = "proto3";

package gomeetfaker.examples;

option go_package = "pb";

import "github.com/gomeet/go-proto-gomeetfaker/gomeetfaker.proto";

message Book {
  string uuid   = 1 [(gomeetfaker.field).uuid.version = "V4"];
  string author = 2 [(gomeetfaker.field).name.name = true];
  string title  = 3 [(gomeetfaker.field).lorem.string = true];
  string isbn10 = 4 [(gomeetfaker.field).code.isbn10 = true];
  string isbn13 = 5 [(gomeetfaker.field).code.isbn13 = true];
}
```

The generated code is understandable. Take a look:

```go
func NewBookGomeetFaker() *Book {
	this := &Book{}
	this.Uuid = uuid.New().String()
	this.Author = faker.Name().Name()
	this.Title = faker.Lorem().String()
	this.Isbn10 = faker.Code().Isbn10()
	this.Isbn13 = faker.Code().Isbn13()
	return this
}
```

For an exhaustive list of faker rules see [examples/full/pb/pb.proto](examples/full/pb/pb.proto)

## Installing and using

The `protoc` compiler expects to find plugins named `proto-gen-XYZ` on the execution `$PATH`. So first:

```sh
export PATH=${PATH}:${GOPATH}/bin
```

Then, do the usual

```sh
go get github.com/gomeet/go-proto-gomeetfaker/protoc-gen-gomeetfaker
```

Your `protoc` builds probably look very simple like:

```sh
protoc  \
	--proto_path=. \
	--go_out=. \
	*.proto
```

That's fine, until you encounter `.proto` includes. Because `go-proto-gomeetfaker` uses annotations inside the `.proto`
files themselves, it's `.proto` definition (and the Google `descriptor.proto` itself) need to on the `protoc` include
path. Hence the above becomes:

```sh
protoc  \
	--proto_path=${GOPATH}/src \
	--proto_path=${GOPATH}/src/github.com/google/protobuf/src \
	--proto_path=. \
	--go_out=. \
	--gomeetfaker_out=. \
	*.proto
```

Or with gogo protobufs:

```sh
protoc  \
	--proto_path=${GOPATH}/src \
	--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. \
	--gogo_out=. \
	--gomeetfaker_out=gogoimport=true:. \
	*.proto
```

Basically the magical incantation (apart from includes) is the `--gomeetfaker_out`. That triggers the
`protoc-gen-gomeetfaker` plugin to generate `mymessage.gomeetfaker.pb.go`. That's it :)

## Make directives

- `make` - Installing all development tools and making specific platform binary (in **_build** directory). It's equivalent to `make build`.
- `make build` - Installing all development tools and making specific platform binary (in **_build** directory).
- `make clean` - Removing all generated files (all tools, all compiled files, generated from proto). It's equivalent to `make tools-clean package-clean proto-clean`.
- `make proto` - Generating files from proto (gomeetfaker.proto only).
- `make proto-examples` - Generating files from proto (gomeetfaker.proto only).
- `make proto-clean` - Clean up generated files from the proto file.
- `make tools` - Installing all development tools.
- `make tools-sync` - Re-Syncronizes development tools.
- `make tools-sync-retool` - Re-Syncronizes development tools (retool only).
- `make tools-sync-protoc` - Re-Syncronizes development tools (protoc only).
- `make tools-upgrade` - Upgrading all development tools.
- `make tools-clean` - Uninstall all development tools.
- `make dep` - Executes the `dep ensure` command.

## License

`go-proto-gomeetfaker` is released under the Apache 2.0 license. See the [LICENSE](LICENSE.txt) file for details.

## TODO

- [ ] proto2 support
- [ ] tests
- [ ] more examples
- [ ] remove [github.com/dmgk/faker](https://github.com/dmgk/faker) depency
