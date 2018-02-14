package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	name    = "gomeet" // injected with -ldflags in Makefile
	version = "latest" // injected with -ldflags in Makefile
)

//
// cliCmd represents the cli command
var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Return version",
		Long: `Use this command

Example :
  $ gomeet version

`,
		Run: versionRun,
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cliCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func versionRun(cmd *cobra.Command, args []string) {
	fmt.Printf("%s version %s", name, version)
}
