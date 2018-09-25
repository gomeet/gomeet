# Set your environment variables here
# for more info see run.sh and run-console.sh
#
#{{ upperSnakeCase .ProjectGroupName }}_PATH="$GOPATH/src/{{ .ProjectGroupGoPkg }}"
#{{ upperSnakeCase .ProjectGroupName }}_EXEC_TYPE="make" # go, make
#{{ upperSnakeCase .ProjectGroupName }}_JWT_SECRET=""
#{{ upperSnakeCase .ProjectGroupName }}_MAX_RECV_MSG_SIZE=10
#{{ upperSnakeCase .ProjectGroupName }}_MAX_SEND_MSG_SIZE=10
#{{ upperSnakeCase .Name }}_ADDRESS=":13000"
{{ range .SubServices -}}{{ $ss := . }}#{{ upperSnakeCase $ss.Name }}_ADDRESS="inprocgrpc"
{{ end -}}
{{ range .QueueTypes }}{{ if eq . "memory" -}}
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_QUEUE_WORKER_COUNT=4
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_QUEUE_MAX_SIZE=100
{{ else if eq . "rabbitmq" }}# rabbitmq support is not yet implemented
{{ else if eq . "zeromq" }}# zeromq support is not yet implemented
{{ else if eq . "sqs" }}# sqs support is not yet implemented
{{ end }}{{ end -}}
{{ range .DbTypes -}}
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_MIGRATE="true"
{{ end -}}
{{ range .SubServices -}}{{ $ss := . }}{{ range .DbTypes -}}
#SVC_{{ upperSnakeCase $ss.ShortName }}_{{ upperSnakeCase . }}_MIGRATE="true"
{{ end -}}{{ end -}}
{{ range .DbTypes }}{{ if eq . "mysql" }}
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_USERNAME="gomeet"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD="toto{{ lower . }}"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_SERVER="localhost"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PORT="3306"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DATABASE="{{ lowerSnakeCase $.Name }}"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DSN="${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_USERNAME:${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD@tcp(${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_SERVER:${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PORT)/${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DATABASE"
{{ else if eq . "postgres" }}
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_USERNAME="gomeet"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD="toto{{ lower . }}"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_SERVER="localhost"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PORT="5432"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DATABASE="{{ lowerSnakeCase $.Name }}"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DSN="host=${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_SERVER port=${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PORT user=${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_USERNAME dbname=${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DATABASE password=${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD"
{{ else if eq . "sqlite" }}
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DSN="/tmp/{{ lowerSnakeCase $.Name }}.sqlite3.db"
{{ else if eq . "mssql" }}
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_USERNAME="gomeet"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD="toto{{ lower . }}"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_SERVER="localhost"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PORT="1433"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DATABASE="{{ lowerSnakeCase $.Name }}"
#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DSN="sqlserver://${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_USERNAME:${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD@${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_SERVER:${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_PORT?database=${{ upperSnakeCase $.Name }}_{{ upperSnakeCase . }}_DB_DATABASE"
{{ end }}
{{ end -}}
{{ range .SubServices }}{{ $ss := . }}{{ range .QueueTypes }}{{ if eq . "memory" }}
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_QUEUE_WORKER_COUNT=4
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_QUEUE_MAX_SIZE=100
{{ else if eq . "rabbitmq" }}# rabbitmq support is not yet implemented
{{ else if eq . "zeromq" }}# zeromq support is not yet implemented
{{ else if eq . "sqs" }}# sqs support is not yet implemented
{{ end }}{{ end }}{{ range .DbTypes }}{{ if eq . "mysql" }}
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_USERNAME="gomeet"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD="toto{{ lower . }}"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_SERVER="localhost"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PORT="3306"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DATABASE="{{ lowerSnakeCase $ss.Name }}"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DSN="${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_USERNAME:${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD@tcp(${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_SERVER:${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PORT)/${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DATABASE"
{{ else if eq . "postgres" }}
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_USERNAME="gomeet"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD="toto{{ lower . }}"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_SERVER="localhost"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PORT="5432"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DATABASE="{{ lowerSnakeCase $ss.Name }}"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DSN="host=${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_SERVER port=${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PORT user=${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_USERNAME dbname=${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DATABASE password=${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD"
{{ else if eq . "sqlite" }}
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DSN="/tmp/{{ lowerSnakeCase $ss.Name }}.sqlite3.db"
{{ else if eq . "mssql" }}
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_USERNAME="gomeet"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD="toto{{ lower . }}"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_SERVER="localhost"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PORT="1433"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DATABASE="{{ lowerSnakeCase $ss.Name }}"
#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DSN="sqlserver://${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_USERNAME:${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PASSWORD@${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_SERVER:${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_PORT?database=${{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase . }}_DB_DATABASE"
{{ end }}{{ end }}{{ end -}}
{{ range .ExtraServeFlags }}#{{ upperSnakeCase $.Name }}_{{ upperSnakeCase .Name }}={{ if eq .Type "string" }}"{{ end }}{{ .DefaultValue }}{{ if eq .Type "string" }}"{{ end }} #{{ $.Name }}: {{ .Description }}
{{ end -}}
{{ range .SubServices }}{{ $ss := . }}{{ range .ExtraServeFlags }}#{{ upperSnakeCase $ss.Name }}_{{ upperSnakeCase .Name }}={{ if eq .Type "string" }}"{{ end }}{{ .DefaultValue }}{{ if eq .Type "string" }}"{{ end }} #{{ $ss.Name }}: {{ .Description }}
{{ end -}}
{{ end }}
