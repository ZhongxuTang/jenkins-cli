package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	appversion "github.com/lemonsoul/jenkins-cli/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the jenkins-cli version information",
	Long:  `Display the version, commit, and build date for the jenkins-cli binary`,
	Run: func(cmd *cobra.Command, args []string) {
		info := appversion.Info()
		fmt.Printf("Version: %s\n", info["version"])
		fmt.Printf("Commit: %s\n", info["commit"])
		fmt.Printf("Build Date: %s\n", info["buildDate"])
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
