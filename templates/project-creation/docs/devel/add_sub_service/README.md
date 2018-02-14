# Sub service declaration

__Todo : automate__

- {{ "{{SubServiceNamePascalCase}}" }} - service name in PascalCase
- {{ "{{SubServiceNameKebabCase}}" }} - service name kebab-case
- {{ "{{SubServiceNameUpperSnakeCase}}" }} - service name in UPPER_SNAKE_CASE
- {{ "{{SubServiceNameLowerSnakeCase}}" }} - service name in lower_snake_case
- {{ "{{CurentServiceNameLowerCamelCase}}" }} - current service name in lowerCamelCase
- {{ "{{SubServiceVersion}}" }} - service version
- {{ "{{SubServiceDevPort}}" }} - port for hack/run.sh

__Warning: keep the comments, thanks__

## 1. Import Client

Search the comment `SUB-SERVICES DEFINITION : import-client` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}Client "{{ .ProjectGroupGoPkg }}/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}/client"
```

## 2. Import Proto

Search the comment `SUB-SERVICES DEFINITION : import-pb` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}Pb "{{ .ProjectGroupGoPkg }}/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}/{{ .GoProtoPkgAlias }}"
```

## 3. Address var

Search the comment `SUB-SERVICES DEFINITION : var-address` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}Address string
```

## 4. Address param

Search the comment `SUB-SERVICES DEFINITION : param-address` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}Address string,
```

## 5. Address flag

Search the comment `SUB-SERVICES DEFINITION : flag-address` and add this code in list

```go
// {{ "{{SubServiceNamePascalCase}}" }} service address
serveCmd.PersistentFlags().StringVar(&svc{{ "{{SubServiceNamePascalCase}}" }}Address, "svc-{{ "{{SubServiceNameKebabCase}}" }}-address", "", "{{ "{{SubServiceNamePascalCase}}" }} service address (host:port)")
```

## 6. Address register

Search the comment `SUB-SERVICES DEFINITION : register-address` and add this code in list

```go
 svc{{ "{{SubServiceNamePascalCase}}" }}Address,
```

## 7. Bind the address to server

Search the comment `SUB-SERVICES DEFINITION : register-address-to-server` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}Address: svc{{ "{{SubServiceNamePascalCase}}" }}Address,
```

## 8. Log the address

Search the comment `SUB-SERVICES DEFINITION : log-address` and add this code in list

```go
"svc{{ "{{SubServiceNamePascalCase}}" }}Address": svc{{ "{{SubServiceNamePascalCase}}" }}Address,
```

## 9. Client var

Search the comment `SUB-SERVICES DEFINITION : client-var` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}GrpcClient svc{{ "{{SubServiceNamePascalCase}}" }}Pb.{{ "{{SubServiceNamePascalCase}}" }}Client
```

## 10. Init function

Search the comment `SUB-SERVICES DEFINITION : func-init` and add this code in list

```go
// init{{ "{{SubServiceNamePascalCase}}" }}Client initializes the gRPC client connecting to the {{ "{{SubServiceNamePascalCase}}" }} service.
func (s *userServer) init{{ "{{SubServiceNamePascalCase}}" }}Client() error {
  if s.svc{{ "{{SubServiceNamePascalCase}}" }}GrpcClient != nil {
    return nil
  }

  cli, err := svc{{ "{{SubServiceNamePascalCase}}" }}Client.NewGomeetClient(s.svc{{ "{{SubServiceNamePascalCase}}" }}Address,
    grpcTimeout,
    s.caCertificate,
    s.certificate,
    s.privateKey,
  )

  if err != nil {
    return err
  }

  s.svc{{ "{{SubServiceNamePascalCase}}" }}GrpcClient = cli.GetGRPCClient()

  return nil
}
```

## 11. ServicesStatus function

Search the comment `SUB-SERVICES DEFINITION : func-status` and add this code in list

```go
func svc{{ "{{SubServiceNamePascalCase}}" }}Status(s *{{ "{{CurentServiceNameLowerCamelCase}}" }}Server, ctx context.Context, svcCtx context.Context) *{{ .GoProtoPkgAlias }}.ServiceStatus {
  if err := s.init{{ "{{SubServiceNamePascalCase}}" }}Client(); err != nil {
    return serviceStatusErrorHandler(ctx, &{{ .GoProtoPkgAlias }}.VersionResponse{"{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}", "unknow"}, err, "Init client error")
  }
  ver, err := s.svc{{ "{{SubServiceNamePascalCase}}" }}GrpcClient.Version(svcCtx, &svc{{ "{{SubServiceNamePascalCase}}" }}Pb.EmptyMessage{})
  if err != nil {
    return serviceStatusErrorHandler(ctx, &{{ .GoProtoPkgAlias }}.VersionResponse{"{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}", "unknow"}, err, "Version call error")
  }
  return serviceStatusSuccessHandler(&{{ .GoProtoPkgAlias }}.VersionResponse{ver.GetName(), ver.GetVersion()})
}
```

## 12. Add call of sub services status call

Search the comment `SUB-SERVICES DEFINITION : call-status` and add this code in list

```go
svc{{ "{{SubServiceNamePascalCase}}" }}Status,
```

## 13. Add tests

__The list is sorted sorted in alphabetical order__

Search the comment `SUB-SERVICES DEFINITION : test-functest` and add this code in list

```go
expected = append(expected, &{{ .GoProtoPkgAlias }}.ServiceStatus{"{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}", "unknow", {{ .GoProtoPkgAlias }}.ServiceStatus_UNAVAILABLE, ""})
```

Search the comment `SUB-SERVICES DEFINITION : test-unit` and add this code in list

```go
expected = append(expected, &{{ .GoProtoPkgAlias }}.ServiceStatus{"{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}", "unknow", {{ .GoProtoPkgAlias }}.ServiceStatus_UNAVAILABLE, ""})
```

## 14. dep

Edit `Gopkg.toml` and add this code

```toml
[[constraint]]
  name = "{{ .ProjectGroupGoPkg }}/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}"
  version = "{{ "{{SubServiceVersion}}" }}"
```

And do

```sh
make dep
```

## 15. hack/run.sh

Search the comment `SUB-SERVICES DEFINITION : run.sh` and add this code in list

```sh
$SCRIPTPATH/../vendor/{{ .ProjectGroupGoPkg }}/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}/_build/packaged/$GOOS-$GOARCH/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }} \
  serve --address ":{{ "{{SubServiceDevPort}}" }}" &
```

And add at the end

```sh
    --svc-{{ "{{SubServiceNameKebabCase}}" }}-address "localhost:{{ "{{SubServiceDevPort}}" }}" \
```

## 16. Update Makefile

Search the make variable `ALL_SVC` and add the service name `{{ "{{SubServiceNameKebabCase}}" }}`

```Makefile
ALL_SVC=... {{ "{{SubServiceNameKebabCase}}" }}
```

Search the comment `SUB-SERVICES DEFINITION : make-tag-docker-compose` and add this code in list

```Makefile
DOCKER_TAG_SVC_{{ "{{SubServiceNameUpperSnakeCase}}" }} = $(shell cat ./vendor/{{ .ProjectGroupGoPkg }}/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}/VERSION | tr +- __)
```

Search the comment `SUB-SERVICES DEFINITION : make-tag-docker-compose-to-env` and add this code in list

```Makefile
echo "TAG_SVC_{{ "{{SubServiceNameUpperSnakeCase}}" }}=$(DOCKER_TAG_SVC_{{ "{{SubServiceNameUpperSnakeCase}}" }})" >> .env
```

## 17. Update docker-compose

Add something like this code to `docker-compose.yml`

```yaml
...
  svc-{{ "{{SubServiceNameKebabCase}}" }}:
    image: {{ .ProjectGroupName }}/{{ .Prefix }}{{ "{{SubServiceNameKebabCase}}" }}:${tag_svc_{{ "{{SubServiceNameLowerSnakeCase}}" }}}
    # deploy:
    #   replicas: 5
    #   resources:
    #     limits:
    #       cpus: "0.1"
    #       memory: 50M
    #   restart_policy:
    #     condition: on-failure
    command: serve
    # ports:
    #   - 13001:13000
    networks:
      - monitoring-back
      - grpc
      - http
...
  svc-{{ "{{CurrentService}}" }}:
    ...
    command: serve --svc-{{ "{{SubServiceNameKebabCase}}" }}-address "svc-{{ "{{SubServiceNameKebabCase}}" }}:13000"
    ...
...
```

# 18. Declare a Prometheus job

Add something like this code to `infra/prometheus/prometheus.yml`

```yaml
scrape_configs:
  - job_name: 'prometheus'
  ...
  static_configs:
    - targets: ... "svc-{{ "{{SubServiceNameKebabCase}}" }}:13000" ...

...
  - job_name: 'svc-{{ "{{SubServiceNameKebabCase}}" }}'
    scrape_interval: 5s
    static_configs:
      - targets:
        - "svc-{{ "{{SubServiceNameKebabCase}}" }}:13000"
...
```
