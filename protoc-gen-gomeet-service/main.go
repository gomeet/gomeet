// Command protoc-gen-gomeet-service is a plugin for Google protocol buffer
// compiler to generate a Gomeet project's microservices, which create a
// the project's gRPC services definitions with it's console, cli client and some sugar.
//
// You rarely need to run this program directly. Instead, put this program
// into your $PATH with a name "protoc-gen-gomeet-service" and run
//   protoc --gomeet-service_out="project_pkg={{ .GoPkg }};default_prefixes={{ .DefaultPrefixes }};sub_services={{ .SubServicesGoPackgeCommaSeparated }}:$GOPATH/src"
//
// See README.md for more details.
package main

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/gomeet/gomeet/utils/project"
)

func main() {
	// Force color output
	if runtime.GOOS != "windows" {
		color.NoColor = false
	}

	g := generator.New()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		g.Error(err, "reading input")
	}

	if err = proto.Unmarshal(data, g.Request); err != nil {
		g.Error(err, "parsing input proto")
	}

	if len(g.Request.FileToGenerate) == 0 {
		g.Fail("no files to generate")
	}

	g.CommandLineParameters(g.Request.GetParameter())
	var (
		subServices     []string
		projectPkg      string
		defaultPrefixes string
		defaultPort     string
		dbTypes         string
		uiType          string
		queueTypes      string
		cronTasks       string
		extraServeFlags string
	)
	useGogoImport := false
	if parameter := g.Request.GetParameter(); parameter != "" {
		for _, param := range strings.Split(parameter, ";") {
			parts := strings.Split(param, "=")
			if len(parts) != 2 {
				log.Printf("warning: invalid parameter: %q", param)
				continue
			}
			switch parts[0] {
			case "sub_services":
				subServices = strings.Split(parts[1], ",")
			case "db_types":
				dbTypes = parts[1]
			case "ui_type":
				uiType = parts[1]
			case "queue_types":
				queueTypes = parts[1]
			case "cron_tasks":
				cronTasks = parts[1]
			case "default_port":
				defaultPort = parts[1]
			case "default_prefixes":
				defaultPrefixes = parts[1]
			case "project_pkg":
				projectPkg = parts[1]
			case "extra_serve_flags":
				extraServeFlags = parts[1]
			case "gogoimport":
				useGogoImport, err = strconv.ParseBool(parts[1])
				if err != nil {
					g.Error(err, "parsing gogoimport option")
				}
			default:
				log.Printf("warning: unknown parameter: %q", param)
			}
		}
	}
	if projectPkg == "" {
		g.Fail("no project_pkg parameter found")
	}

	p, err := project.New(projectPkg)
	if err != nil {
		g.Error(err, "project initialization fail")
	}
	p.SetDefaultPrefixes(defaultPrefixes)
	p.SetDefaultPort(defaultPort)
	p.SetSubServices(subServices)

	if dbTypes != "" {
		err := p.SetDbTypes(dbTypes)
		if err != nil {
			g.Error(err, "bad db_types parameter")
		}
	}

	if uiType != "" {
		err := p.SetUiType(uiType)
		if err != nil {
			g.Error(err, "bad ui_type parameter")
		}
	}

	if queueTypes != "" {
		err := p.SetQueueTypes(queueTypes)
		if err != nil {
			g.Error(err, "bad queue_types parameter")
		}
	}

	if cronTasks != "" {
		err := p.SetCronTasks(cronTasks)
		if err != nil {
			g.Error(err, "bad cron_tasks parameter")
		}
	}

	if extraServeFlags != "" {
		err := p.SetExtraServeFlags(extraServeFlags)
		if err != nil {
			g.Error(err, "bad extra_serve_flags parameter")
		}
	}

	p.UseGogoGen(useGogoImport)

	if err := p.GenFromProto(g.Request); err != nil {
		g.Error(err, "project template generation fail")
	}
}
