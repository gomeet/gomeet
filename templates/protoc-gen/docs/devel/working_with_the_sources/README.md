# Working with the sources

## Install

### Install requirements

{{ .Name }} needs some requirements :

- [golang](https://golang.org/doc/install)
- [protobuf](https://github.com/google/protobuf)
- [git flow](https://danielkummer.github.io/git-flow-cheatsheet/)
- [docker](https://docs.docker.com/engine/installation/)
- [docker-compose](https://docs.docker.com/compose/install/)
- [Unzip](http://www.info-zip.org/UnZip.html)
{{ if .HasUi }}{{ if .HasUiElm }}- [NodeJS](https://guide.elm-lang.org/install.html)
- [yarn package manager](https://yarnpkg.com/en/docs/install)
{{ end }}{{ end }}{{ range .DbTypes }}{{ if eq . "mysql" }}- [{{ lower . }}](https://www.{{ lower . }}.com/) or [mariaDB clone](https://mariadb.com/)
{{ else if eq . "postgres" }}- [postgreSQL](https://www.postgresql.org/){{ if $.HasPostgis }}
- [postGIS](http://postgis.net/docs/manual-2.4/){{ end }}
{{ else if eq . "sqlite" }}- [sqlite3](https://www.sqlite.org/)
{{ else if eq . "mssql" }}- [sql-server](https://www.microsoft.com/fr-fr/sql-server/sql-server-2016)
{{ end }}{{ end }}
#### On Linux (Ubuntu Xenial)

```bash
sudo apt-get update
sudo apt-get install -y build-essential git software-properties-common python-software-properties
sudo add-apt-repository -y ppa:longsleep/golang-backports
sudo add-apt-repository -y ppa:maarten-fonville/protobuf
sudo apt-get update
sudo apt-get install -y golang-go protobuf-compiler git-flow
sudo apt-get install unzip

echo -e "export GOPATH=\$(go env GOPATH)\nexport PATH=\${PATH}:\${GOPATH}/bin" >> ~/.bashrc
source ~/.bashrc

# docker install cf. https://docs.docker.com/engine/installation/linux/docker-ce/ubuntu/
# 1. Add repository
sudo apt-get update
sudo apt-get install apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo apt-key fingerprint 0EBFCD88
# Response is something like this
#   pub   4096R/0EBFCD88 2017-02-22
#         Key fingerprint = 9DC8 5822 9FC7 DD38 854A  E2D8 8D81 803C 0EBF CD88
#   uid                  Docker Release (CE deb) <docker@docker.com>
#   sub   4096R/F273FCD8 2017-02-22
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update
sudo apt-get install docker-ce

# docker compose install https://docs.docker.com/compose/install/#install-compose
# 1. download the latest version
sudo curl \
     -L https://github.com/docker/compose/releases/download/1.17.0/docker-compose-`uname -s`-`uname -m` \
     -o /usr/local/bin/docker-compose
# 2. Apply executable permissions to the binary
sudo chmod +x /usr/local/bin/docker-compose
{{ if .HasUi }}{{ if .UiType eq "elm" }}
# 3. Install ui tools chain
# install nodejs
curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -
sudo apt-get install -y nodejs
# install elm-platform
npm install -g elm

# 4. Install databases
{{ end }}{{ else }}
# 3. Install databases
{{ end }}
{{ if .DbTypes }}{{ range .DbTypes }}{{ if eq . "mysql" }}
# install mariadb/mysql
sudo apt-get install mariadb-server
{{ else if eq . "postgres" }}{{ if $.HasPostgis }}# install postgreSQL/PostGIS
sudo apt-get install postgresql-10-postgis-2.4 postgresql-contrib
{{ else }}# install postgreSQL
sudo apt-get install postgresql postgresql-contrib{{ end }}
{{ else if eq . "sqlite" }}
# install sqlite3
sudo apt-get install sqlite3
{{ else if eq . "mssql" }}
# install sql-server
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | sudo apt-key add -
sudo add-apt-repository "$(wget -qO- https://packages.microsoft.com/config/ubuntu/16.04/mssql-server-2017.list)"
sudo apt-get update
sudo apt-get install -y mssql-serve mssql-tools unixodbc-devr
sudo /opt/mssql/bin/mssql-conf setup
echo 'export PATH="$PATH:/opt/mssql-tools/bin"' >> ~/.bash_profile
echo 'export PATH="$PATH:/opt/mssql-tools/bin"' >> ~/.bashrc
source ~/.bashrc{{ end }}{{ end }}{{ end }}
```

#### On MacOSX

```bash

brew install go
brew install git
brew install protobuf
brew install git-flow-avh

echo -e "export GOPATH=\$(go env GOPATH)\nexport PATH=\${PATH}:\${GOPATH}/bin" >> ~/.bashrc
echo "" >> ~/.bashrc
source ~/.bashrc

{{ if .HasUi }}{{ if .UiType eq "elm" }}
# for ui
# install nodejs
brew install node
# install elm-platform
npm install -g elm
{{ end }}{{ end }}

# for docker see https://docs.docker.com/docker-for-mac/install/
{{ if .DbTypes }}{{ range .DbTypes }}{{ if eq . "mysql" }}brew install mysql
{{ else if eq . "postgres" }}brew install postgresql{{ if $.HasPostgis }}
brew install postgis{{ end }}
{{ else if eq . "sqlite" }}brew install sqlite
{{ else if eq . "mssql" }}# use Docker to run sql-server see https://docs.microsoft.com/fr-fr/sql/linux/quickstart-install-connect-docker
sudo docker pull microsoft/mssql-server-linux:2017-latest
sudo docker run -e 'ACCEPT_EULA=Y' -e 'MSSQL_SA_PASSWORD=<YourStrong!Passw0rd>' \
   -p 1401:1433 --name sql1 \
   -d microsoft/mssql-server-linux:2017-latest{{ end }}{{ end }}{{ end }}
```

#### On Windows

```txt
TODO
```

### Check version dependencies

```bash
$ protoc --version
libprotoc 3.3.0

$ go version
go version go1.8.1 ...snip...

$ echo $GOPATH
...snip...

$ docker --version
Docker version 17.06.2-ce, build cec0b72

$ docker-compse --version
docker-compose version 1.17.0, build ac53b73
```

### Install and build from source

```bash
# clone repository in $GOPATH
mkdir -p $GOPATH/src/{{ .ProjectGroupGoPkg }}
cd $GOPATH/src/{{ .ProjectGroupGoPkg }}
git clone https://{{ .GoPkg }}.git
cd {{ .Name }}

# initalize git-flow needed for make release
git checkout master
git checkout develop
git flow init -d

# use make
make
_build/{{ .Name }}
```

{{ if .DbTypes }}### Database initialization
{{ range .DbTypes }}{{ if eq . "mysql" }}
- MySQL database

```bash
$ sudo mysql -u root -p
Enter password: ***********
MariaDB [(none)]> CREATE DATABASE {{ lowerSnakeCase $.Name }};
...
MariaDB [(none)]> CREATE DATABASE {{ lowerSnakeCase $.Name }}_test;
...
MariaDB [(none)]> GRANT ALL PRIVILEGES ON {{ lowerSnakeCase $.Name }}.* To '<USERNAME>'@'localhost' IDENTIFIED BY '<PASSWORD>';
...
MariaDB [(none)]> GRANT ALL PRIVILEGES ON {{ lowerSnakeCase $.Name }}_test.* To '<USERNAME>'@'localhost' IDENTIFIED BY '<PASSWORD>';
...
```
{{ else if eq . "postgres" }}
- postgreSQL database

```bash
$ sudo su - postgres
createuser -d -E -i -l -P -r -s <USERNAME>

$ sudo -u postgres createdb -O <USERNAME> {{ lowerSnakeCase $.Name }}
$ sudo -u postgres createdb -O <USERNAME> {{ lowerSnakeCase $.Name }}_test
```{{ if $.HasPostgis }}

- postGIS

```bash
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }} -c "CREATE EXTENSION postgis;"
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }} -c "CREATE EXTENSION postgis_topology;"
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }} -c "CREATE EXTENSION fuzzystrmatch;"
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }} -c "CREATE EXTENSION postgis_tiger_geocoder;"

$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }}_test -c "CREATE EXTENSION postgis;"
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }}_test -c "CREATE EXTENSION postgis_topology;"
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }}_test -c "CREATE EXTENSION fuzzystrmatch;"
$ sudo -u postgres psql -d {{ lowerSnakeCase $.Name }}_test -c "CREATE EXTENSION postgis_tiger_geocoder;"
```{{ end }}

{{ else if eq . "sqlite" }}
- sqlite3 database

```bash
TODO in github.com/gomeet/gomeet/templates/protoc-gen/docs/devel/working_with_the_sources/README.md
```
{{ else if eq . "mssql" }}
- sql-server database

```bash
TODO in github.com/gomeet/gomeet/templates/protoc-gen/docs/devel/working_with_the_sources/README.md
```{{ end }}{{ end }}{{ end }}
### Make directives

- `make` - Installing all development tools and making specific platform binary. It's equivalent to `make build`
- `make build` - Installing all development tools and making specific platform binary
- `make clean` - Removing all generated files (all tools, all compiled files, generated from proto). It's equivalent to `make tools-clean package-clean proto-clean`
- `make docker-test` - Executing `make test` inside docker (not yet ready it's doesn't run very well)
- `make docker` - Building docker image
- `make docker-push` - Push the docker image to docker registry server - default registry is `docker.io` it can be overloaded via the environment variable `DOCKER_REGISTRY` like this `DOCKER_REGISTRY={{ "{{hostname}}" }}:{{ "{{port}}" }} make docker-push`
- `make install` - Performing a `go install` command
- `make package-arm32` - build linux arm32 packages
- `make package-arm64` - build linux arm64  packages
- `make package-amd64-linux` - build linux amd64 packages
- `make package-amd64-darwin` - build darwin amd64 packages
- `make package-amd64-openbsd` - build OpenBSD packages
- `make package-amd64-windows` - build windows amd64 packages
- `make package-amd64` - build all amf64 packages - alias of `make package-amd64-linux package-amd64-darwin package-amd64-openbsd package-amd64-windows`
- `make package` - Building all packages (multi platform, and docker image)
- `make package-clean` - Clean up the builded packages
- `make package-proto` - Building the `_build/packaged/proto.tgz` file with dirstribluables protobuf files
{{ if .HasUi }}- `make ui` - Generation of a virtual file system that is compiled with the binary from files inside `ui/assets`
{{ end }}- `make proto` - Generating files from proto
- `make proto-clean` - Clean up generated files from the proto file
- `make release` - Making a release (see below)
- `make start` - Building docker image and performing a `docker-compose up -d` command
- `make stop` - Performing a `docker-compose down` command
- `make test` - Runing tests (shortcut for `make test-funcs test-unit test-race`)
- `make test-func` - Runing functionals tests
- `make test-unit` - Runing units tests
- `make test-race` - Runing units tests with race trace
- `make tools` - Installing all development tools
- `make tools-sync` - Re-Syncronizes development tools
- `make tools-sync-retool` - Re-Syncronizes `retool` tool
- `make tools-sync-protoc` - Re-Syncronizes `protoc` tool
- `make tools-upgrade-gomeet` - Upgrading all gomeet's development tools [gomeet-tools-markdown-server](github.com/gomeet/gomeet-tools-markdown-server), [protoc-gen-gomeetfaker](github.com/gomeet/go-proto-gomeetfaker/protoc-gen-gomeetfaker), [gomeet & protoc-gen-gomeet-service](https://github.com/gomeet/gomeet)
- `make tools-upgrade` - Upgrading all development tools
- `make tools-clean` - Uninstall all development tools
- `make dep` - Executes the `dep ensure` command
- `make dep-init` - Executes the `dep init` command (normaly never)
- `make dep-prune` - Executes the `dep prune` command
- `make dep-update-{{ .ProjectGroupName }} [individual svc name without {{ .Prefix }} prefix|default all]` - Executes the `dep ensure -update {{ .ProjectGroupGoPkg }}/{{ .Prefix }}[individual svc name without {{ .Prefix }} prefix|default all]`
- `make dep-update-gomeet-utils` - Executes the `dep ensure -update github.com/gomeet/gomeet`
- `make doc-server` - Run a markdown documentation server
- `make run` - Run the server (via hack/run.sh script)
- `make run-console` - Run the console (via hack/run-console.sh script)
- `make run-dev` - Run the server with hot compile (via hack/run-dev.sh script)
- `make gomeet-regenerate-project` - regenerate the project with [gomeet](https://github.com/gomeet/gomeet) be careful this replaces files except for the protobuf file

#### Add a tool

Build tool chain:

```shell
make tools-sync
make tools
```

Add a tool dependency:

```shell
_tools/bin/retool add retool github.com/jteeuwen/go-bindata/go-bindata origin/master
```

Use a tool:

```shell
_tools/bin/go-bindata
```

Commit changes

#### Make a release

```bash
make release <Git flow option : start|finish> <Release version : major|minor|patch> [Release version metadata (optional)]
```

- Git flow option
  - The `start` option does not finish the `git flow release` so to finish the release and prepare `VERSION` file in `develop` branch do :
  ```bash
  git flow release finish "v$(cat VERSION)"
  # NB: _tools/bin/semver is compiled with "make tools"
  NEW_DEV_VERSION=`_tools/bin/semver -patch -build "dev" $(cat VERSION)` && echo "$NEW_DEV_VERSION" > VERSION
  git add VERSION
  git commit -m "Bump version - v$(cat VERSION)"
  git push --tag
  git push origin develop
  git push origin master
  ```
  - The `finish` option does it for you.

- Release version and metadata (if `VERSION` file in `develop` branch is `1.1.1+dev`) :
  - `make release start patch` start and publish the `release/1.1.1` git flow release branch
  - `make release start patch rc.1` start and publish the `release/1.1.1+rc.1` git flow release branch
  - `make release start minor` start and publish the `release/1.2.0` git flow release branch
  - `make release start minor foo.1` start and publish the `release/1.2.0+foo.1` git flow release branch
  - `make release start major` start and publish the `release/2.0.0` git flow release branch
  - `make release start major foo.1` start and publish the `release/2.0.0+foo.1` git flow release branch
  - `make release finish patch` make the `1.1.1` release and `VERSION` file in `develop` branch is `1.1.2+dev`
  - `make release finish patch rc.1` make the `1.1.1+rc.1` release and `VERSION` file in `develop` branch branch is `1.1.2+dev`
  - `make release finish minor` make the `1.2.0` release and `VERSION` file in `develop` branch is `1.2.1+dev`
  - `make release finish minor foo.1` make the `1.2.0+foo.1` release and `VERSION` file in `develop` branch is `1.2.0+dev`
  - `make release finish major` make the `2.0.0` release and `VERSION` file in `develop` branch is `2.0.1+dev`
  - `make release finish major foo.1` make the `2.0.0+foo.1` release and `VERSION` file in `develop` branch is `2.0.1+dev`

#### Manual steps

```bash
make tools
NEW_VERSION="x.y.z" && \
  git flow release start "v$NEW_VERSION" && \
  echo $NEW_VERSION > VERSION
git add VERSION
git commit -m "Bump version - v$(cat VERSION)"
awk \
  -v \
  log_title="## Unreleased\n\n- Nothing\n\n## $(cat VERSION) - $(date +%Y-%m-%d)" \
  '{gsub(/## Unreleased/,log_title)}1' \
  CHANGELOG.md > CHANGELOG.md.tmp && \
    mv CHANGELOG.md.tmp CHANGELOG.md
git add CHANGELOG.md
git commit -m "Improved CHANGELOG.md"
make package
git add _build/packaged/
git commit -m "Added v$(cat VERSION) packages"
git flow release publish "v$(cat VERSION)"
git flow release finish "v$(cat VERSION)"
# NB: _tools/bin/semver is compiled with "make tools"
NEW_DEV_VERSION=`_tools/bin/semver -patch -build "dev" $(cat VERSION)` && \
  echo $NEW_DEV_VERSION > VERSION
git add VERSION
git commit -m "Bump version - v$(cat VERSION)"
git push --tag
git push origin develop
git push origin master
```

## Use docker (no requirement)

- See gomeet/gomeet-builder docker image ([Docker Hub](https://hub.docker.com/r/gomeet/gomeet-builder/) - [Source](https://github.com/gomeet/gomeet-builder)).

## Working with gotools

If {{ .Name }} repository is private and you use [Gogs](https://gogs.io/) has remote server.

To work with go tools (`go get` et `dep`) it's necesary to configure gogs, git and ssh.

1. Add your ssh key to your gogs user settings https://<GOGS_ADDRESS>/user/settings/ssh

2. In your local git config (`~/.gitconfig`) add these lines :

```
...
[url "ssh://<GOGS_SSH_USER>@<GOGS_ADDRESS>:<GOGS_SSH_PORT (default: 10022)>"]
	insteadOf = https://<GOGS_ADDRESS>
...
```

The SSH URL might require a trailing slash depending on the version of Git (observed on 2.9.3).

3. In your local ssh config (`~/.ssh/config`) add these lines :

```
...
Host <GOGS_ADDRESS>
  HostName <GOGS_ADDRESS>
  Port <GOGS_SSH_PORT (default: 10022)>
  User <GOGS_SSH_USER>
...
```

__WARNING__ : be sure that `ssh-agent` is running

## Uninstall

### Remove source

```bash
rm $GOPATH/bin/{{ .Name }}
rm -rf $GOPATH/src/{{ .ProjectGroupGoPkg }}
```

### Uninstall dependencies

- On Linux (Ubuntu Xenial)

```bash
sudo apt-get autoremove --purge golang-go protobuf-compiler git-flow
sudo add-apt-repository -r --purge ppa:longsleep/golang-backports
sudo add-apt-repository -r ppa:maarten-fonville/protobuf
sudo rm /etc/apt/sources.list.d/longsleep-ubuntu-golang-backports-xenial.list*
sudo rm /etc/apt/sources.list.d/maarten-fonville-ubuntu-protobuf-xenial.list*
sudo apt-get autoremove --purge build-essential git software-properties-common python-software-properties
sudo apt-get update

sed -i.bak ':a;N;$!ba;s/\nexport GOPATH=$(go env GOPATH)\nexport PATH=\${PATH}:\${GOPATH}\/bin//g' ~/.bashrc
unset GOPATH
source ~/.bashrc

# uninstall docker
sudo apt-get purge docker-ce
sudo rm -rf /var/lib/docker
sudo add-apt-repository -r "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-key del "Docker Release (CE deb) <docker@docker.com>"
```

- On MacOSX

```bash
brew rm go
brew rm git
brew rm protobuf
brew rm git-flow-avh

sed -i.bak '/export GOPATH=\$(go env GOPATH)/d' ~/.bashrc
unset GOPATH
source ~/.bashrc
```

- On Windows

```txt
TODO
```

## Some usual procedures

- To add a {{ upperPascalCase .ProjectGroupName }} subservice as dependency see [this](../add_sub_service/README.md)
- To add a new gRPC service see [this](../add_grpc_service/README.md)
