package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"

	"github.com/gomeet/gomeet/utils/project"
)

type askOptions int

const (
	DEFAULT_NO askOptions = iota
	DEFAULT_YES
)

type FinishMsgParams struct {
	Path string
	Arg  string
	Cmd  []string
}

const finishMsg = `  $ cd {{ .Path }}
{{ range .Cmd }}  $ {{ . }}
{{ end }}
`

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new microservice",
	Run:   new,
}

var (
	subService      string
	defaultPrefixes string
	protoAlias      string
	force           bool
	silent          bool
	noGogo          bool
	dbTypes         string
	uiType          string
	queueTypes      string
	cronTasks       string
	extraServeFlags string
	defaultPort     string
	out             = colorable.NewColorableStdout()
)

func init() {
	newCmd.PersistentFlags().StringVar(&subService, "sub-services", "", "Sub services dependencies (comma separated)")
	newCmd.PersistentFlags().StringVar(&defaultPrefixes, "default-prefixes", project.GomeetDefaultPrefixes(), "List of prefixes (comma separated)")
	newCmd.PersistentFlags().StringVar(&defaultPort, "default-port", project.DefaultRawPort(), "Default port")
	newCmd.PersistentFlags().StringVar(&protoAlias, "proto-alias", project.DEFAULT_PROTO_PKG_ALIAS, "Protobuf pakage alias")
	newCmd.PersistentFlags().BoolVar(&force, "force", false, "Replace files if exists")
	newCmd.PersistentFlags().BoolVar(&noGogo, "no-gogo", false, "if is true the protoc plugin is protoc-gen-go else it's protoc-gen-gogo in the Makefile file")
	newCmd.PersistentFlags().StringVar(&dbTypes, "db-types", "", fmt.Sprintf("DB types [%s] (comma separated)", strings.Join(project.GomeetAllowedDbTypes(), ",")))
	newCmd.PersistentFlags().StringVar(&uiType, "ui-type", "none", fmt.Sprintf("UI type [%s]", strings.Join(project.GomeetAllowedUiTypes(), "|")))
	newCmd.PersistentFlags().StringVar(&queueTypes, "queue-types", "", fmt.Sprintf("Queue types [%s] (comma separated)", strings.Join(project.GomeetAllowedQueueTypes(), ",")))
	newCmd.PersistentFlags().StringVar(&extraServeFlags, "extra-serve-flags", "", "extra serve flags passed to gRPC server format [<name-of-flag>@<type-of-flag[string|int]>|<flag description (no comma, no semicolon, no colon)>|<default value>] (comma separated)")
	newCmd.PersistentFlags().StringVar(&cronTasks, "cron-tasks", "", "Cron tasks (comma separated)")
	newCmd.PersistentFlags().BoolVarP(&silent, "silent", "s", false, "Silent questions with default answers")

	rootCmd.AddCommand(newCmd)
}

func new(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("You must supply a path for the service, e.g gomeet new github.com/gomeet/gomeet-svc-myservice\n")
		return
	}

	if os.Getenv("CI") != "" {
		silent = true
	}

	name := args[0]
	p, err := project.New(name)
	if err != nil {
		er(err)
	}

	fmt.Printf("Creating project in %s\n", p.Path())
	if !silent && !askIsOK("Is this OK?", DEFAULT_NO) {
		fmt.Println("Exiting..")
		return
	}

	if subService != "" {
		subServices := strings.Split(subService, ",")
		err := p.SetSubServices(subServices)
		if err != nil {
			er(err)
		}
	}

	if defaultPrefixes != "" {
		err := p.SetDefaultPrefixes(defaultPrefixes)
		if err != nil {
			er(err)
		}
	}

	if defaultPort != "" {
		_, err := p.SetDefaultPort(defaultPort)
		if err != nil {
			er(err)
		}
	}

	if dbTypes != "" {
		err := p.SetDbTypes(dbTypes)
		if err != nil {
			er(err)
		}
	}

	if uiType != "" {
		err := p.SetUiType(uiType)
		if err != nil {
			er(err)
		}
	}

	if queueTypes != "" {
		err := p.SetQueueTypes(queueTypes)
		if err != nil {
			er(err)
		}
	}

	if cronTasks != "" {
		err := p.SetCronTasks(cronTasks)
		if err != nil {
			er(err)
		}
	}

	if extraServeFlags != "" {
		err := p.SetExtraServeFlags(extraServeFlags)
		if err != nil {
			er(err)
		}
	}

	if protoAlias != "" {
		p.SetDefaultProtoPkgAlias(protoAlias)
	}

	keepProtoModel := true
	if !silent && force {
		keepProtoModel = !askIsOK("Keep the protobuf, models, tools, package, ui files, auth_and_acl_funcs and helpers files ?", DEFAULT_YES)
	}
	p.UseGogoGen(!noGogo)

	// create new project
	err = p.ProjectCreation(!force, keepProtoModel)
	if err != nil {
		er(err)
	}

	if force {
		return
	}

	if silent || askIsOK("Print tree?", DEFAULT_NO) {
		p.PrintTreeFolder()
	}

	// Create a new template and parse the finishMsg into it.
	t := template.Must(template.New("finishMsg").Parse(finishMsg))

	// git init and ...
	fmt.Println("To finish project initialization do :")
	err = t.Execute(
		os.Stdout,
		FinishMsgParams{
			Path: p.Path(),
			Arg:  name,
			Cmd:  p.AfterProjectCreationCmd(),
		},
	)
	if err != nil {
		er(err)
	}

	if !silent && !askIsOK("Do it?", DEFAULT_NO) {
		fmt.Println("Exiting..")
		return
	}

	if err := p.ExecAfterProjectCreationCmd(true); err != nil {
		er(err)
	}

	// git flow init -d and ...
	fmt.Println("")
	fmt.Println("To git flow initialization do :")
	err = t.Execute(
		os.Stdout,
		FinishMsgParams{
			Path: p.Path(),
			Arg:  name,
			Cmd:  p.AfterProjectCreationGitFlowCmd(),
		},
	)
	if err != nil {
		er(err)
	}

	if !silent && !askIsOK("Do it?", DEFAULT_NO) {
		fmt.Println("Exiting..")
		return
	}

	if err := p.ExecAfterProjectCreationGitFlowCmd(true); err != nil {
		er(err)
	}
}

func askIsOK(msg string, defaultVal askOptions) bool {
	var (
		yTxt, nTxt string
		checkVal   string
	)

	switch defaultVal {
	case DEFAULT_NO:
		checkVal, yTxt, nTxt = "y", "[y]", "[N]"
	case DEFAULT_YES:
		checkVal, yTxt, nTxt = "n", "[Y]", "[n]"
	}

	if msg == "" {
		msg = "Is this OK?"
	}

	fmt.Fprintf(out, "\n%s\n%ses/%so\n",
		msg,
		color.YellowString(yTxt),
		color.CyanString(nTxt),
	)

	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	return strings.Contains(strings.ToLower(scan.Text()), checkVal)
}

func er(err error) {
	if err != nil {
		fmt.Fprintf(out, "%s: %s \n",
			color.RedString("[ERROR]"),
			err.Error(),
		)
		os.Exit(-1)
	}
}
