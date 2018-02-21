package project

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	ggdescriptor "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"

	"github.com/gomeet/gomeet/utils/project/helpers"
	tmplHelpers "github.com/gomeet/gomeet/utils/project/templates/helpers"
)

const DEFAULT_PROTO_PKG_ALIAS = "pb"

var allowedDbTypes = []string{"mysql", "postgres", "sqllite", "mssql"}

func GomeetDefaultPrefixes() string {
	return helpers.GomeetDefaultPrefixes
}

func GomeetAllowedDbTypes() []string {
	return allowedDbTypes
}

type Project struct {
	*helpers.PkgNfo

	SubServices          []*helpers.PkgNfo
	dbTypes              []string
	folder               *folder
	protoRegistry        *ggdescriptor.Registry
	protoFiles           []*descriptor.FileDescriptorProto
	defaultProtoPkgAlias *string
	isGogoGen            bool
}

func New(inputPath string) (*Project, error) {
	path, err := helpers.Path(inputPath)
	if err != nil {
		return nil, err
	}
	goPkg := helpers.Base(path)

	pkgNfo, err := helpers.NewPkgNfo(goPkg, "")
	if err != nil {
		return nil, err
	}
	p := &Project{pkgNfo, nil, []string{}, nil, nil, nil, nil, false}
	p.SetDefaultPrefixes("")
	p.SetDefaultProtoPkgAlias("")
	return p, nil
}

func (p Project) PrettyPrint() {
	b, _ := json.MarshalIndent(p, "", "  ")
	fmt.Println(string(b))
}

func (p *Project) SetDefaultPrefixes(s string) error {
	p.PkgNfo.SetDefaultPrefixes(s)

	if len(p.SubServices) > 0 {
		for _, ss := range p.SubServices {
			ss.SetDefaultPrefixes(p.DefaultPrefixes())
		}
	}

	return nil
}
func (p *Project) SetDbTypes(s string) error {
	if s != "" {
		dbTypes := strings.Split(s, ",")
		for _, dbType := range dbTypes {
			dbType = strings.ToLower(strings.TrimSpace(dbType))
			ok := false
			for _, allowedDbType := range GomeetAllowedDbTypes() {
				if dbType == allowedDbType {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("%s isn't allowed dbType", dbType)
			}
			p.dbTypes = append(p.dbTypes, dbType)
		}
	}

	return nil
}

func (p *Project) SetDefaultProtoPkgAlias(s string) error {
	if s == "" {
		s = DEFAULT_PROTO_PKG_ALIAS
	}
	p.defaultProtoPkgAlias = &s

	return nil
}

func (p Project) PrintTreeFolder()                              { p.folder.print() }
func (p Project) GomeetPkg() string                             { return helpers.GomeetPkg() }
func (p Project) IsGogoGen() bool                               { return p.isGogoGen }
func (p Project) GomeetGeneratorUrl() string                    { return "https://" + p.GomeetPkg() }
func (p Project) ProtoFiles() []*descriptor.FileDescriptorProto { return p.protoFiles }
func (p Project) DbTypes() []string                             { return p.dbTypes }

func (p *Project) UseGogoGen(b bool) {
	p.isGogoGen = b
}

func (p *Project) SetSubServices(subServices []string) error {
	if len(subServices) > 0 {
		for _, subSvcPkg := range subServices {
			subSvcPkg = strings.TrimSpace(subSvcPkg)
			if subSvcPkg == "" {
				continue
			}
			//ss, err := sub_service.NewSubSevice(subSvcPkg, p.DefaultPrefixes())
			pkgNfo, err := helpers.NewPkgNfo(subSvcPkg, p.DefaultPrefixes())
			if err != nil {
				return err
			}
			p.SubServices = append(p.SubServices, pkgNfo)
		}
	}

	return nil
}

func (p *Project) setProjectCreationTree(keepFile, keepProtoModel bool) (err error) {
	f := newFolder(p.Name(), p.Path())
	f.addTree(".", "project-creation", nil, keepFile)
	pbFolder := f.getFolder("pb")
	modelsFolder := f.getFolder("models")

	// reset "pb" folder if proto alias isn't "pb"
	protoAlias, err := p.GoProtoPkgAlias()
	if err != nil {
		return err
	}
	if protoAlias != "pb" {
		pbFolder, err = f.addTree(protoAlias, "project-creation/pb", nil, keepFile)
		if err != nil {
			return err
		}
		f.delete("pb")
	}

	pbFile, err := pbFolder.getFile("proto.proto")
	if err != nil {
		return err
	}
	pbFile.KeepIfExists = keepProtoModel

	modelsFile, err := modelsFolder.getFile("models.go")
	if err != nil {
		return err
	}
	modelsFile.KeepIfExists = keepProtoModel

	// rename generic proto.proto to <short project name>.proto
	err = pbFolder.renameFile(
		"proto.proto",
		fmt.Sprintf("%s.proto", p.ShortName()),
	)
	if err != nil {
		return err
	}

	// if not use gogo no need gogo's descriptors as third party
	if !p.IsGogoGen() {
		f.delete("third_party/github.com/gogo")
	}

	p.folder = f

	return nil
}

func (p *Project) ProjectCreation(keepFile, keepProtoModel bool) error {
	if err := p.setProjectCreationTree(keepFile, keepProtoModel); err != nil {
		return err
	}
	if _, err := os.Stat(p.Path()); os.IsNotExist(err) {
		err = os.MkdirAll(p.Path(), os.ModePerm)
		if err != nil {
			return err
		}
	}
	err := p.folder.render(*p)
	if err != nil {
		return nil
	}

	err = filepath.Walk(p.Path()+"/hack/", func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chmod(name, 0755)
		}
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (p Project) AfterProjectCreationCmd() (r []string) {
	r = append(r, "git init")
	r = append(r, "git add .")
	r = append(r, fmt.Sprintf("git commit -m 'First commit (gomeet new %s)'", p.GoPkg()))
	r = append(r, "make tools-sync proto dep dep-prune test")
	r = append(r, "git add .")
	r = append(r, "git commit -m 'Added tools and dependencies'")
	return r
}

func (p Project) AfterProjectCreationGitFlowCmd() (r []string) {
	r = append(r, "git flow init -d")
	return r
}

func (p Project) ExecAfterProjectCreationCmd(v bool) error {
	return p.execCommands(v, p.AfterProjectCreationCmd())
}

func (p Project) ExecAfterProjectCreationGitFlowCmd(v bool) error {
	return p.execCommands(v, p.AfterProjectCreationGitFlowCmd())
}

func (p Project) execCommands(v bool, cmds []string) error {
	if len(cmds) < 1 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(cmds))
	for _, sCmd := range cmds {
		var err error
		parts := helpers.ParseCmd(sCmd)
		cmdName := parts[0]
		cmdArgs := parts[1:]
		cmd := exec.Command(cmdName, cmdArgs...)
		cmd.Dir = p.Path()
		if v {
			// verbose
			fmt.Printf("%s $ %s\n", color.CyanString(p.Path()), sCmd)
			cmdReader, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
				return err
			}

			scanner := bufio.NewScanner(cmdReader)
			go func() {
				defer wg.Done()
				for scanner.Scan() {
					fmt.Println(scanner.Text())
				}
			}()

			err = cmd.Start()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
				return err
			}

			err = cmd.Wait()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
				return err
			}
		} else {
			wg.Done()
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Project) setProtoRegistry(req *plugin.CodeGeneratorRequest) error {
	registry := ggdescriptor.NewRegistry()
	if err := registry.Load(req); err != nil {
		return err
	}
	p.protoRegistry = registry
	tmplHelpers.SetRegistry(p.protoRegistry)
	for _, file := range req.GetProtoFile() {
		if _, err := p.protoRegistry.LookupFile(file.GetName()); err != nil {
			return fmt.Errorf("registry: failed to lookup file %q -- %s", file.GetName(), err)
		}
		if file.GetName() == req.FileToGenerate[0] {
			p.protoFiles = append(p.protoFiles, file)
		}
	}

	return nil
}

func (p Project) DefaultProtoPkgAlias() string {
	return *p.defaultProtoPkgAlias
}

func (p Project) GoProtoPkgAlias() (string, error) {
	if len(p.ProtoFiles()) > 0 {
		for _, file := range p.ProtoFiles() {
			desc, err := p.protoRegistry.LookupFile(file.GetName())
			if err != nil {
				return "", fmt.Errorf("registry: failed to lookup file %q -- %s", file.GetName(), err)
			}

			return desc.GoPkg.Name, nil
		}
	}

	return p.DefaultProtoPkgAlias(), nil
}

func (p *Project) GenFromProto(req *plugin.CodeGeneratorRequest) error {
	err := p.setProtoRegistry(req)
	if err != nil {
		return err
	}

	f := newFolder(p.Name(), p.Path())
	cmd := f.addFolder("cmd")
	cmd.addFile("cli.go", "protoc-gen/cmd/cli.go.tmpl", nil, false)
	cmd.addFile("root.go", "protoc-gen/cmd/root.go.tmpl", nil, false)
	cmd.addFile("serve.go", "protoc-gen/cmd/serve.go.tmpl", nil, false)
	cmd.addFile("functest.go", "protoc-gen/cmd/functest.go.tmpl", nil, false)
	cmd.addFile("migrate.go", "protoc-gen/cmd/migrate.go.tmpl", nil, false)
	functest := cmd.addFolder("functest")
	functest.addFile("http_metrics.go", "protoc-gen/cmd/functest/http_metrics.go.tmpl", nil, false)
	functest.addFile("types.go", "protoc-gen/cmd/functest/types.go.tmpl", nil, false)
	f.addTree("client", "protoc-gen/client", nil, false)
	f.addTree("docs", "protoc-gen/docs", nil, false)
	rcli := cmd.addFolder("remotecli")
	rcli.addFile("cmd_help.go", "protoc-gen/cmd/remotecli/cmd_help.go.tmpl", nil, false)
	rcli.addFile("remotecli.go", "protoc-gen/cmd/remotecli/remotecli.go.tmpl", nil, false)
	f.addTree("models", "protoc-gen/models", nil, false)
	svc := f.addFolder("service")
	svc.addFile("grpc.go", "protoc-gen/service/grpc.go.tmpl", nil, false)
	svc.addFile("http.go", "protoc-gen/service/http.go.tmpl", nil, false)
	svc.addFile("service.go", "protoc-gen/service/service.go.tmpl", nil, false)
	svc.addFile("service_test.go", "protoc-gen/service/service_test.go.tmpl", nil, false)
	svc.addFile("init_subservice_clients.go", "protoc-gen/service/init_subservice_clients.go.tmpl", nil, false)
	svc.addFile("init_databases.go", "protoc-gen/service/init_databases.go.tmpl", nil, false)
	f.addTree("infra", "protoc-gen/infra", nil, false)
	protoPkg, err := p.GoProtoPkgAlias()
	if err != nil {
		return err
	}
	f.addTree(protoPkg, "protoc-gen/pb", nil, false)
	f.addFile("docker-compose.yml", "protoc-gen/docker-compose.yml.tmpl", nil, false)

	var hasVersion, hasServicesStatus bool
	for _, file := range p.ProtoFiles() {
		if _, err := p.protoRegistry.LookupFile(file.GetName()); err != nil {
			return fmt.Errorf("registry: failed to lookup file %q -- %s", file.GetName(), err)
		}
		for _, service := range file.GetService() {
			for _, method := range service.GetMethod() {
				var (
					fName, tName string
					keepFile     bool
				)
				grpcM := &grpcMethod{
					File:    file,
					Service: service,
					Method:  method,
				}
				switch method.GetName() {
				case "Version":
					// TODO check request/response validity
					hasVersion, fName, tName, keepFile = true, "version", "version", false
					svc.addFile("grpc_version_test.go", "protoc-gen/service/grpc_version_test.go.tmpl", grpcM, false)

				case "ServicesStatus":
					// TODO check request/response validity
					hasServicesStatus, fName, tName, keepFile = true, "services_status", "services_status", false
					svc.addFile("grpc_services_status_test.go", "protoc-gen/service/grpc_services_status_test.go.tmpl", grpcM, false)

				default:
					tName = fmt.Sprintf("%s_%s", streamFromBool(method.GetClientStreaming()), streamFromBool(method.GetServerStreaming()))
					fName = tmplHelpers.LowerSnakeCase(method.GetName())
					keepFile = true
				}
				svc.addFile(fmt.Sprintf("grpc_%s.go", fName), fmt.Sprintf("protoc-gen/service/grpc_%s.go.tmpl", tName), grpcM, keepFile)
				svc.addFile(fmt.Sprintf("grpc_%s_test.go", fName), fmt.Sprintf("protoc-gen/service/grpc_%s_test.go.tmpl", tName), grpcM, keepFile)
				rcli.addFile(fmt.Sprintf("cmd_%s.go", fName), fmt.Sprintf("protoc-gen/cmd/remotecli/cmd_%s.go.tmpl", tName), grpcM, keepFile)
				functest.addFile(fmt.Sprintf("helpers_%s.go", fName), fmt.Sprintf("protoc-gen/cmd/functest/helpers_%s.go.tmpl", tName), grpcM, keepFile)
				functest.addFile(fmt.Sprintf("grpc_%s.go", fName), fmt.Sprintf("protoc-gen/cmd/functest/grpc_%s.go.tmpl", tName), grpcM, false)
				functest.addFile(fmt.Sprintf("http_%s.go", fName), fmt.Sprintf("protoc-gen/cmd/functest/http_%s.go.tmpl", tName), grpcM, false)
			}
		}
	}
	if !hasVersion {
		helpers.Log(
			helpers.LogDangerous,
			"doesn't have Version grpc method this is a part of gomeet service\n",
		)
	}
	if !hasServicesStatus {
		helpers.Log(
			helpers.LogDangerous,
			"doesn't have ServicesStatus grpc method this is a part of gomeet service\n",
		)
	}

	p.folder = f
	return p.folder.render(*p)
}

func streamFromBool(streaming bool) string {
	if streaming {
		return "stream"
	}

	return "unary"
}
