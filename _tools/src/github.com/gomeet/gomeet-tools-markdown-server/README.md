# gomeet-tools-markdown-server

[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

`gomeet-tools-markdown-server` is a simple tool to serve mardown inside gomeet service repository.

## Install

```shell
go install github.com/gomeet/gomeet-tools-markdown-server
```

## Usage

```shell
Usage: gomeet-tools-markdown-server [-v] [--root=DIR] [ADDR]

Options:
  -h --help        Show this screen.
     --version     Show version.
  -v --verbose     Show more information.
     --root=DIR    Document root. [Default: .]
```

Inside a gomeet service

```shell
make run-doc-server
```

## License

`go-proto-gomeetfaker` is released under the Apache 2.0 license. See the [LICENSE](LICENSE.txt) file for details.

## Todo

- [ ] Black friday v2
- [ ] logrus for log
- [ ] Blackfirday todo list extension
- [ ] gitbook menu
