package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/stefancocora/vaultguard/pkg/version"
)

// test in v0.0.1-dev-86e986f+UNCOMMITEDCHANGES
// vaultguard --config tmp/config.yaml version

var bldctxPtr bool

func init() {
	RootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&bldctxPtr, "buildcontext", "b", false, "additionally prints the app  build context")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Print the current version number of %v", version.BinaryName),
	// Short: "Print the current version number",
	Long: `Print the current version number`,
	Run: func(cmd *cobra.Command, args []string) {
		versionCommandParser()
	},
}

func versionCommandParser() {
	vers, err := version.Printvers()
	if err != nil {
		log.Fatal("unable to get version")
	}
	bldctx, err := version.BuildContextCli()
	if err != nil {
		log.Fatal("unable to get cli build context")
	}

	if bldctxPtr {
		fmt.Printf("%v version %s\n", version.BinaryName, vers)
		fmt.Printf("%s\n", bldctx)
		fmt.Println()
		return
	}
	fmt.Printf("%v version %s\n", version.BinaryName, vers)
	fmt.Println()
}
