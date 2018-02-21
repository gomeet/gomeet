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
	protoName       string
	force           bool
	noGogo          bool
	dbTypes         string
	out             = colorable.NewColorableStdout()
)

func init() {
	newCmd.PersistentFlags().StringVar(&subService, "sub-services", "", "Sub services dependencies (comma separated)")
	newCmd.PersistentFlags().StringVar(&defaultPrefixes, "default-prefixes", "", fmt.Sprintf("List of prefixes [%s] (comma separated) - Overloaded with $GOMEET_DEFAULT_PREFIXES", project.GomeetDefaultPrefixes()))
	newCmd.PersistentFlags().StringVar(&protoName, "proto-name", "", "Protobuf pakage name (inside project)")
	newCmd.PersistentFlags().BoolVar(&force, "force", false, "Replace files if exists")
	newCmd.PersistentFlags().BoolVar(&noGogo, "no-gogo", false, "if is true the protoc plugin is protoc-gen-go else it's protoc-gen-gogo in the Makefile file")
	newCmd.PersistentFlags().StringVar(&dbTypes, "db-types", "", fmt.Sprintf("DB types [%s] (comma separated)", strings.Join(project.GomeetAllowedDbTypes(), ",")))

	RootCmd.AddCommand(newCmd)
}

func new(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("You must supply a path for the service, e.g gomeet new github.com/gomeet/gomeet-svc-myservice\n")
		return
	}

	name := args[0]
	p, err := project.New(name)
	if err != nil {
		er(err)
	}

	fmt.Printf("Creating project in %s\n", p.Path())
	if !askIsOK("Is this OK?") {
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

	if dbTypes != "" {
		err := p.SetDbTypes(dbTypes)
		if err != nil {
			er(err)
		}
	}

	if protoName != "" {
		p.SetDefaultProtoPkgAlias(protoName)
	}

	keepProtoModel := true
	if force {
		keepProtoModel = !askIsOK("Are you sure you want to overwrite the protobuf and models files ?")
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

	if askIsOK("Print tree?") {
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

	if !askIsOK("Do it?") {
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

	if !askIsOK("Do it?") {
		fmt.Println("Exiting..")
		return
	}

	if err := p.ExecAfterProjectCreationGitFlowCmd(true); err != nil {
		er(err)
	}
}

func askIsOK(msg string) bool {
	if os.Getenv("CI") != "" {
		return true
	}

	if msg == "" {
		msg = "Is this OK?"
	}

	fmt.Fprintf(out, "%s %ses/%so\n",
		msg,
		color.YellowString("[y]"),
		color.CyanString("[N]"),
	)

	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	return strings.Contains(strings.ToLower(scan.Text()), "y")
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
