package project

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	ggdescriptor "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"

	"github.com/gomeet/gomeet/utils/project/helpers"
	tmplHelpers "github.com/gomeet/gomeet/utils/project/templates/helpers"
)

const (
	DEFAULT_PROTO_PKG_ALIAS = "pb"
	DEFAULT_ELM             = "elm-bulma"
	DEFAULT_PORT            = 50051
)

var (
	allowedDbTypes    = []string{"mysql", "postgres", "postgis", "sqlite", "mssql"}
	allowedUiTypes    = []string{"none", "simple", "elm", "elm-bulma", "elm-milligram", "elm-minimal", "elm-minimal-http"} //TODO "react", "vuejs", ....
	allowedQueueTypes = []string{"memory", "rabbitmq", "zeromq", "sqs"}
)

func GomeetDefaultPrefixes() string {
	return helpers.GomeetDefaultPrefixes
}

func DefaultRawPort() string {
	return fmt.Sprintf("%d", DEFAULT_PORT)
}

func GomeetAllowedDbTypes() []string { return allowedDbTypes }
func GomeetAllowedUiTypes() []string { return allowedUiTypes }

func GomeetAllowedQueueTypes() []string {
	return allowedQueueTypes
}

type serveFlag struct {
	Name         string
	Description  string
	DefaultValue string
	Type         string
}

type Project struct {
	*helpers.PkgNfo

	SubServices          map[string]*Project
	dbTypes              []string
	uiType               string
	queueTypes           []string
	cronTasks            []string
	hasPostgis           bool
	extraServeFlags      []*serveFlag
	folder               *folder
	protoRegistry        *ggdescriptor.Registry
	protoFiles           []*descriptor.FileDescriptorProto
	defaultProtoPkgAlias *string
	isGogoGen            bool
	version              string
	defaultPort          int
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
	//p := &Project{pkgNfo, nil, []string{}, "none", []string{}, []string{}, false, nil, nil, nil, nil, nil, false, "0.0.1+dev"}
	p := &Project{
		PkgNfo:               pkgNfo,
		SubServices:          nil,
		dbTypes:              []string{},
		uiType:               "none",
		queueTypes:           []string{},
		cronTasks:            []string{},
		hasPostgis:           false,
		extraServeFlags:      nil,
		folder:               nil,
		protoRegistry:        nil,
		protoFiles:           nil,
		defaultProtoPkgAlias: nil,
		isGogoGen:            false,
		version:              "0.0.1+dev",
	}
	p.SetDefaultPrefixes("")
	p.SetDefaultProtoPkgAlias(DEFAULT_PROTO_PKG_ALIAS)
	p.SetDefaultPort("")
	return p, nil
}

func (p *Project) SetVersion(v string) {
	p.version = v
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

func extraFlag(myFlag string) (*serveFlag, error) {
	if myFlag == "" {
		return nil, nil
	}
	part := strings.Split(myFlag, "|")
	if len(part) < 1 {
		return nil, errors.New("bad extra serve flags parameter")
	}
	name, description, defaultValue, typ := part[0], "", "", "string"
	if len(part) > 1 {
		description = part[1]
	}
	if len(part) > 2 {
		defaultValue = part[2]
	}
	namePart := strings.Split(name, "@")
	if len(namePart) > 1 {
		name = namePart[0]
		typ = strings.ToLower(namePart[1])
	}

	return &serveFlag{
		Name:         name,
		Description:  description,
		DefaultValue: defaultValue,
		Type:         typ,
	}, nil

}

func (p *Project) SetExtraServeFlags(s string) error {
	if s != "" {
		allFlags := strings.Split(s, ",")
		for _, myFlag := range allFlags {
			if myFlag == "" {
				continue
			}
			aServeFlag, err := extraFlag(myFlag)
			if err != nil {
				return err
			}
			if aServeFlag == nil {
				continue
			}
			p.extraServeFlags = append(p.extraServeFlags, aServeFlag)
		}
	}

	return nil
}

func (p *Project) SetDbTypes(s string) error {
	if s != "" {
		dbTypes := strings.Split(s, ",")
		pgFind := 0
		for _, dbType := range dbTypes {
			if dbType == "postgres" || dbType == "postgis" {
				pgFind++
			}
		}
		if pgFind > 1 {
			return errors.New(`"postgis" is a specal alias for "postgres" these dbTypes aren't allowed together`)
		}
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
			if dbType == "postgis" {
				dbType = "postgres"
				p.hasPostgis = true
			}
			p.dbTypes = append(p.dbTypes, dbType)
		}
	}

	return nil
}

func (p *Project) SetUiType(s string) error {
	p.uiType = "none"
	if s != "" {
		uiType := strings.ToLower(strings.TrimSpace(s))
		ok := false
		for _, allowedUiType := range GomeetAllowedUiTypes() {
			if uiType == allowedUiType {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("\"%s\" isn't allowed in ui_type - allowed ui_type : [%s]", uiType, strings.Join(GomeetAllowedUiTypes(), "|"))
		}
		if uiType == "elm" {
			uiType = DEFAULT_ELM
		}
		p.uiType = uiType
	}

	return nil
}

func (p *Project) SetQueueTypes(s string) error {
	if s != "" {
		queueTypes := strings.Split(s, ",")
		for _, queueType := range queueTypes {
			queueType = strings.ToLower(strings.TrimSpace(queueType))
			ok := false
			for _, allowedQueueType := range GomeetAllowedQueueTypes() {
				if queueType == allowedQueueType {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("%s isn't allowed queueType", queueType)
			}
			p.queueTypes = append(p.queueTypes, queueType)
		}
	}

	return nil
}

func (p *Project) SetCronTasks(s string) error {
	if s != "" {
		cronTasks := strings.Split(s, ",")
		for _, cronTask := range cronTasks {
			cronTask = tmplHelpers.LowerSnakeCase(strings.TrimSpace(cronTask))
			p.cronTasks = append(p.cronTasks, cronTask)
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
func (p Project) GomeetRetoolRev() string                       { return helpers.GomeetRetoolRev }
func (p Project) IsGogoGen() bool                               { return p.isGogoGen }
func (p Project) GomeetGeneratorUrl() string                    { return "https://" + p.GomeetPkg() }
func (p Project) Version() string                               { return p.version }
func (p Project) ProtoFiles() []*descriptor.FileDescriptorProto { return p.protoFiles }
func (p Project) DbTypes() []string                             { return p.dbTypes }
func (p Project) UiType() string                                { return p.uiType }
func (p Project) QueueTypes() []string                          { return p.queueTypes }
func (p Project) CronTasks() []string                           { return p.cronTasks }
func (p Project) ExtraServeFlags() []*serveFlag                 { return p.extraServeFlags }
func (p Project) DefaultPort() int                              { return p.defaultPort }

func (p Project) GoCGOEnabled() int {
	ret := 0
	if len(p.DbTypes()) > 0 {
		for _, dbT := range p.DbTypes() {
			if dbT == "sqlite" {
				ret = 1
				break
			}
		}
	}

	return ret
}

func (p Project) HasDb() bool {
	if len(p.DbTypes()) > 0 {
		return true
	}
	for _, ss := range p.SubServices {
		if ss.HasDb() {
			return true
		}
	}

	return false
}

func (p Project) HasMySqlDb() bool {
	dbTypes := p.DbTypes()
	if len(dbTypes) > 0 {
		for _, dbType := range dbTypes {
			if dbType == "mysql" {
				return true
			}
		}
	}

	for _, ss := range p.SubServices {
		if ss.HasMySqlDb() {
			return true
		}
	}

	return false
}

func (p Project) HasUi() bool {
	uiT := p.UiType()
	return (uiT != "" && uiT != "none")
}

func (p Project) HasUiElm() bool {
	uiT := p.UiType()
	return strings.HasPrefix(uiT, "elm")
}

func (p Project) HasMemoryQueue() bool {
	if len(p.QueueTypes()) > 0 {
		for _, queueType := range p.QueueTypes() {
			if queueType == "memory" {
				return true
			}
		}
	}
	for _, ss := range p.SubServices {
		if ss.HasMemoryQueue() {
			return true
		}
	}

	return false
}

func (p Project) HasPostgis() bool {
	dbT := p.DbTypes()
	if len(dbT) < 1 {
		return false
	}
	for _, ss := range p.SubServices {
		if ss.HasPostgis() {
			return true
		}
	}

	for _, typ := range p.DbTypes() {
		if typ == "postgres" {
			return p.hasPostgis
		}
	}

	return false
}

func (p *Project) UseGogoGen(b bool) {
	p.isGogoGen = b
}

func parseSubService(s string) []string {
	r := regexp.MustCompile(`'.*?'|\[.*?\]|\S+`)
	res := r.FindAllString(s, -1)
	for k, v := range res {
		mod := strings.Trim(v, " ")
		mod = strings.Trim(mod, "[")
		mod = strings.Trim(mod, `]`)
		mod = strings.Trim(mod, " ")

		res[k] = mod
	}
	return res
}

func (p Project) CountSubServices() int {
	return len(p.SubServices)
}

func (p Project) CountSubServicesWithDbTypes() int {
	r := 0
	if len(p.SubServices) > 0 && p.HasDb() {
		for _, ss := range p.SubServices {
			if ss.HasDb() {
				r++
			}
		}
	}
	if len(p.DbTypes()) > 0 {
		r++
	}
	return r
}

func (p *Project) SubServicesMonolithHelp() string {
	servicesKeys := []string{}
	for k := range p.SubServices {
		servicesKeys = append(servicesKeys, k)
	}
	sort.Strings(servicesKeys)

	ssStrings := []string{}
	for _, k := range servicesKeys {
		ss := p.SubServices[k]
		ssString := fmt.Sprintf(
			"  - if \"svc-%s-address\" is empty or is equal to \"inprocgrpc\" the",
			tmplHelpers.LowerKebabCase(ss.ShortName()),
		)
		ssFlags := []string{}
		if len(ss.DbTypes()) > 0 {
			for _, dbType := range ss.DbTypes() {
				ssFlags = append(
					ssFlags,
					fmt.Sprintf(
						"\"svc-%s-%s-dsn\"",
						ss.ShortName(),
						strings.ToLower(dbType),
					),
				)
				ssFlags = append(
					ssFlags,
					fmt.Sprintf(
						"\"svc-%s-%s-migrate\"",
						ss.ShortName(),
						strings.ToLower(dbType),
					),
				)
			}
		}
		if len(ss.QueueTypes()) > 0 {
			for _, queueType := range ss.QueueTypes() {
				switch queueType {
				case "memory":
					ssFlags = append(
						ssFlags,
						fmt.Sprintf(
							"\"svc-%s-%s-queue-worker-count\"",
							ss.ShortName(),
							strings.ToLower(queueType),
						),
					)
					ssFlags = append(
						ssFlags,
						fmt.Sprintf(
							"\"svc-%s-%s-queue-max-size\"",
							ss.ShortName(),
							strings.ToLower(queueType),
						),
					)
				case "rabbitmq", "zeromq", "sqs":
				// rabbitmq, zeromq, sqs support is not yet implemented
				default:
					// FIXME what to do?
				}
			}
		}
		if len(ss.ExtraServeFlags()) > 0 {
			for _, ssf := range ss.ExtraServeFlags() {
				ssFlags = append(
					ssFlags,
					fmt.Sprintf(
						"\"svc-%s-%s\"",
						tmplHelpers.LowerKebabCase(ss.ShortName()),
						tmplHelpers.LowerKebabCase(ssf.Name),
					),
				)
			}
		}
		ssString = fmt.Sprintf(
			"%s %s",
			ssString,
			strings.Join(ssFlags, ", "),
		)
		if len(ss.DbTypes()) > 0 || len(ss.QueueTypes()) > 0 || len(ss.ExtraServeFlags()) > 0 {
			ssString = fmt.Sprintf(
				"%s flags are used to launch \"svc-%s\" server in the main process",
				ssString,
				tmplHelpers.LowerKebabCase(ss.ShortName()),
			)
		} else {
			ssString = fmt.Sprintf(
				"%s\"svc-%s\" server is launched in the main process",
				ssString,
				tmplHelpers.LowerKebabCase(ss.ShortName()),
			)
		}
		ssStrings = append(ssStrings, ssString)
	}
	return strings.Join(ssStrings, "\n")
}

func (p *Project) SubServicesDef() string {
	return p.subServicesDef(false)
}
func (p *Project) SubServicesDefMultiline() string {
	return p.subServicesDef(true)
}

func (p *Project) subServicesDef(multiline bool) string {
	servicesKeys := []string{}
	for k := range p.SubServices {
		servicesKeys = append(servicesKeys, k)
	}
	sort.Strings(servicesKeys)

	ssStrings := []string{}
	for _, k := range servicesKeys {
		ss := p.SubServices[k]
		ssString := fmt.Sprintf(
			"%s[version@%s]",
			ss.GoPkg(),
			ss.Version(),
		)
		ssString = fmt.Sprintf(
			"%s[default_port@%d]",
			ssString,
			ss.DefaultPort(),
		)
		if len(ss.DbTypes()) > 0 {
			ssString = fmt.Sprintf(
				"%s[db_types@%s]",
				ssString,
				strings.Join(ss.DbTypes(), "|"),
			)
		}
		if len(ss.QueueTypes()) > 0 {
			ssString = fmt.Sprintf(
				"%s[queue_types@%s]",
				ssString,
				strings.Join(ss.QueueTypes(), "|"),
			)
		}
		if len(ss.SubServices) > 0 {
			sssPkg := []string{}
			for _, sss := range ss.SubServices {
				sssPkg = append(sssPkg, sss.GoPkg())
			}
			ssString = fmt.Sprintf(
				"%s[sub_services@%s]",
				ssString,
				strings.Join(sssPkg, "|"),
			)
		}
		if len(ss.ExtraServeFlags()) > 0 {
			for _, ssf := range ss.ExtraServeFlags() {
				ssString = fmt.Sprintf(
					"%s[%s@%s|%s|%s]",
					ssString,
					ssf.Name,
					ssf.Type,
					ssf.Description,
					ssf.DefaultValue,
				)
			}
		}
		ssStrings = append(ssStrings, ssString)
	}
	sep := ","
	if multiline {
		sep = `,\
`
	}
	return strings.Join(ssStrings, sep)
}

func (p *Project) SetSubServices(subServices []string) error {
	if len(subServices) > 0 {
		neededSubServices := []string{}
		dependenciesTree := map[string][]string{}
		p.SubServices = make(map[string]*Project)
		iPort := p.DefaultPort()
		for _, subSvcPkg := range subServices {
			part := strings.Split(subSvcPkg, "[")
			if len(part) < 1 {
				continue
			}
			subSvcPkg = strings.TrimSpace(part[0])
			if subSvcPkg == "" {
				continue
			}

			subSvc, err := New(subSvcPkg)
			if err != nil {
				return err
			}
			subSvc.SetDefaultPrefixes(p.DefaultPrefixes())
			ssParams := parseSubService("[" + strings.Join(part[1:], "["))

			var (
				ssVersion     string
				ssDefaultPort string
				ssDbTypes     []string
				ssQueueTypes  []string
				ssFlags       []string
				ssSubServices []string
			)
			for _, ssFlag := range ssParams {
				switch {
				case strings.HasPrefix(ssFlag, "db_types@"):
					ssDbTypes = strings.Split(strings.TrimPrefix(ssFlag, "db_types@"), "|")
				case strings.HasPrefix(ssFlag, "queue_types@"):
					ssQueueTypes = strings.Split(strings.TrimPrefix(ssFlag, "queue_types@"), "|")
				case strings.HasPrefix(ssFlag, "version@"):
					ssVersion = strings.TrimPrefix(ssFlag, "version@")
				case strings.HasPrefix(ssFlag, "default_port@"):
					ssDefaultPort = strings.TrimPrefix(ssFlag, "default_port@")
				case strings.HasPrefix(ssFlag, "sub_services@"):
					ssSubServices = strings.Split(strings.TrimPrefix(ssFlag, "sub_services@"), "|")
					dependenciesTree[subSvc.GoPkg()] = ssSubServices
				default:
					ssFlags = append(ssFlags, ssFlag)
				}
			}
			if ssVersion != "" {
				subSvc.SetVersion(ssVersion)
			}
			if ssDefaultPort != "" {
				iPort, err = subSvc.SetDefaultPort(ssDefaultPort)
				if err != nil {
					return err
				}
			}
			if len(ssFlags) > 0 {
				subSvc.SetExtraServeFlags(strings.Join(ssFlags, ","))
			}
			if len(ssDbTypes) > 0 {
				subSvc.SetDbTypes(strings.Join(ssDbTypes, ","))
			}
			if len(ssQueueTypes) > 0 {
				subSvc.SetQueueTypes(strings.Join(ssQueueTypes, ","))
			}
			if len(ssSubServices) > 0 {
				neededSubServices = append(neededSubServices, ssSubServices...)
			}

			p.SubServices[subSvc.GoPkg()] = subSvc
		}
		for _, needed := range neededSubServices {
			if _, ok := p.SubServices[needed]; !ok {
				subSvc, err := New(needed)
				if err != nil {
					return err
				}
				subSvc.SetDefaultPrefixes(p.DefaultPrefixes())
				iPortB := iPort + 10
				iPort, err = subSvc.SetDefaultPort(fmt.Sprintf("%d", iPortB))
				p.SubServices[subSvc.GoPkg()] = subSvc
			}
		}

		for pkg, deps := range dependenciesTree {
			for _, dep := range deps {
				if p.SubServices[pkg].SubServices == nil {
					p.SubServices[pkg].SubServices = make(map[string]*Project)
				}
				p.SubServices[pkg].SubServices[dep] = p.SubServices[dep]
			}
		}
	}

	return nil
}

func (p *Project) setProjectCreationTree(keepFile, keepProtoModelUi bool) (err error) {
	f := newFolder(p.Name(), p.Path())
	f.addTree(".", "project-creation", nil, keepFile)
	pbFolder := f.getFolder("pb")

	// keep some files if needed
	keepFiles := map[string][]*folder{
		"tools.json":            []*folder{f},
		"Gopkg.toml":            []*folder{f},
		"CHANGELOG.md":          []*folder{f},
		"VERSION":               []*folder{f},
		".directory":            []*folder{f},
		"proto.proto":           []*folder{pbFolder},
		"run-env.sh":            []*folder{f.getFolder("hack")},
		"models.go":             []*folder{f.getFolder("models")},
		"auth_and_acl_funcs.go": []*folder{f.getFolder("service")},
		"helpers.go":            []*folder{f.getFolder("cmd").getFolder("remotecli"), f.getFolder("service")},
		"icon.png":              []*folder{f.getFolder(".metadata")},
	}
	for myFileName, myFolders := range keepFiles {
		for _, myFolder := range myFolders {
			myFile, err := myFolder.getFile(myFileName)
			if err != nil {
				return err
			}
			myFile.KeepIfExists = keepProtoModelUi
		}
	}

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

	f.delete("ui")
	if p.HasUi() {
		_, err := f.addTree("ui", fmt.Sprintf("project-creation/ui/%s", p.UiType()), nil, keepProtoModelUi)
		if err != nil {
			return err
		}
	}

	p.folder = f

	return nil
}

func (p *Project) ProjectCreation(keepFile, keepProtoModelUi bool) error {
	if err := p.setProjectCreationTree(keepFile, keepProtoModelUi); err != nil {
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
	//r = append(r, "make tools-sync proto dep dep-prune test")
	r = append(r, "make tools-sync proto dep")
	r = append(r, "git add .")
	r = append(r, "git commit -m 'Added tools and dependencies'")
	switch {
	case p.HasUi() && p.HasUiElm():
		r = append(r, "make ui-setup")
		r = append(r, "make ui")
		r = append(r, "git add .")
		r = append(r, "git commit -m 'Added ui'")
	}

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

func (p *Project) SetDefaultPort(rawPort string) (int, error) {
	rawPort = strings.Trim(rawPort, " ")
	if len(rawPort) == 0 {
		rawPort = DefaultRawPort()
	}

	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("Invalid default-port number : %s", err)
	}

	iPort := int(port)
	p.defaultPort = iPort

	if len(p.SubServices) > 0 {
		for _, ss := range p.SubServices {
			iPortB := iPort + 10
			iPort, err = ss.SetDefaultPort(fmt.Sprintf("%d", iPortB))
			if err != nil {
				return iPortB, err
			}
		}
	}

	return iPort, nil
}

func (p Project) DefaultProtoPkgAlias() string {
	return *p.defaultProtoPkgAlias
}

func (p Project) GoProtoPkgName() (string, error) {
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

func (p Project) GoProtoPkgAlias() (string, error) {
	if len(p.ProtoFiles()) > 0 {
		for _, file := range p.ProtoFiles() {
			desc, err := p.protoRegistry.LookupFile(file.GetName())
			if err != nil {
				return "", fmt.Errorf("registry: failed to lookup file %q -- %s", file.GetName(), err)
			}
			var alias string
			if desc.GoPkg.Alias != "" {
				alias = desc.GoPkg.Alias
			} else {
				alias = desc.GoPkg.Name
			}
			return alias, nil
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
	cmd.addFile("helpers_flags.go", "protoc-gen/cmd/helpers_flags.go.tmpl", nil, false)
	cmd.addFile("helpers_config.go", "protoc-gen/cmd/helpers_config.go.tmpl", nil, false)
	cmd.addFile("cli.go", "protoc-gen/cmd/cli.go.tmpl", nil, false)
	cmd.addFile("root.go", "protoc-gen/cmd/root.go.tmpl", nil, false)
	cmd.addFile("serve.go", "protoc-gen/cmd/serve.go.tmpl", nil, false)
	cmd.addFile("functest.go", "protoc-gen/cmd/functest.go.tmpl", nil, false)
	cmd.addFile("migrate.go", "protoc-gen/cmd/migrate.go.tmpl", nil, false)
	cmd.addFile("crontask.go", "protoc-gen/cmd/crontask.go.tmpl", nil, false)
	crontaskFolder := cmd.addFolder("crontask")
	crontaskFolder.addFile("clients.go", "protoc-gen/cmd/crontask/clients.go.tmpl", nil, false)
	if len(p.CronTasks()) > 0 {
		for _, crontask := range p.CronTasks() {
			// we fake a gRPC method to passe it to the template
			grpcM := &grpcMethod{Method: &descriptor.MethodDescriptorProto{Name: &crontask}}
			// we keep the file if exists
			crontaskFolder.addFile(fmt.Sprintf("%s.go", crontask), "protoc-gen/cmd/crontask/crontask.go.tmpl", grpcM, true)
		}
	}
	functest := cmd.addFolder("functest")
	functest.addFile("http_metrics.go", "protoc-gen/cmd/functest/http_metrics.go.tmpl", nil, false)
	functest.addFile("types.go", "protoc-gen/cmd/functest/types.go.tmpl", nil, false)
	f.addTree("client", "protoc-gen/client", nil, false)
	f.addTree("docs", "protoc-gen/docs", nil, false)
	rcli := cmd.addFolder("remotecli")
	rcli.addFile("cmd_help.go", "protoc-gen/cmd/remotecli/cmd_help.go.tmpl", nil, false)
	rcli.addFile("remotecli.go", "protoc-gen/cmd/remotecli/remotecli.go.tmpl", nil, false)
	f.addTree("models", "protoc-gen/models", nil, false)
	srv := f.addFolder("server")
	srv.addFile("server.go", "protoc-gen/server/server.go.tmpl", nil, false)
	svc := f.addFolder("service")
	svc.addFile("grpc.go", "protoc-gen/service/grpc.go.tmpl", nil, false)
	svc.addFile("service.go", "protoc-gen/service/service.go.tmpl", nil, false)
	svc.addFile("service_test.go", "protoc-gen/service/service_test.go.tmpl", nil, false)
	svc.addFile("init_subservice_clients.go", "protoc-gen/service/init_subservice_clients.go.tmpl", nil, false)
	svc.addFile("init_databases.go", "protoc-gen/service/init_databases.go.tmpl", nil, false)
	svc.addFile("init_queues.go", "protoc-gen/service/init_queues.go.tmpl", nil, false)
	f.addTree("infra", "protoc-gen/infra", nil, false)
	f.addTree("hack", "protoc-gen/hack", nil, false)
	// added pb directory
	f.addTree("pb", "protoc-gen/pb", nil, false)
	f.addFile("docker-compose.yml", "protoc-gen/docker-compose.yml.tmpl", nil, false)
	f.addFile(".travis.yml", "protoc-gen/.travis.yml.tmpl", nil, false)

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
	if err = p.folder.render(*p); err != nil {
		return err
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

func streamFromBool(streaming bool) string {
	if streaming {
		return "stream"
	}

	return "unary"
}
